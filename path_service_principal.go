package vspheresecrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/vmware/govmomi"
)

const (
	SecretTypeSP       = "service_principal"
	SecretTypeStaticSP = "static_service_principal"
)

// SPs will be created with a far-future expiration in Azure... relvant for vSphere?
var spExpiration = 10 * 365 * 24 * time.Hour

func secretServicePrincipal(b *vsphereSecretBackend) *framework.Secret {
	return &framework.Secret{
		Type:   SecretTypeSP,
		Renew:  b.spRenew,
		Revoke: b.spRevoke,
	}
}

func secretStaticServicePrincipal(b *vsphereSecretBackend) *framework.Secret {
	return &framework.Secret{
		Type:   SecretTypeStaticSP,
		Renew:  b.spRenew,
		Revoke: b.staticSPRevoke,
	}
}

func pathServicePrincipal(b *vsphereSecretBackend) *framework.Path {
	return &framework.Path{
		Pattern: fmt.Sprintf("session/%s", framework.GenericNameRegex("role")),
		Fields: map[string]*framework.FieldSchema{
			"role": {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the Vault role",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathSPRead,
		},
		HelpSynopsis:    pathServicePrincipalHelpSyn,
		HelpDescription: pathServicePrincipalHelpDesc,
	}
}

// pathSPRead generates Azure credentials based on the role credential type.
func (b *vsphereSecretBackend) pathSPRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleName := d.Get("role").(string)

	role, err := getRole(ctx, roleName, req.Storage)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return logical.ErrorResponse(fmt.Sprintf("role '%s' does not exist", roleName)), nil
	}

	var resp *logical.Response

	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if role.Password != "" {
		resp, err = b.createStaticSPSecret(ctx, client, roleName, role)
	} else {
		resp, err = b.createSPSecret(ctx, client, roleName, role)
	}

	if err != nil {
		return nil, err
	}

	resp.Secret.TTL = role.TTL
	resp.Secret.MaxTTL = role.MaxTTL
	return resp, nil
}

// createSPSecret generates a new App/Service Principal.
func (b *vsphereSecretBackend) createSPSecret(ctx context.Context, c *client, roleName string, role *roleEntry) (*logical.Response, error) {
	/*
		// Create the App, which is the top level object to be tracked in the secret
		// and deleted upon revocation. If any subsequent step fails, the App is deleted.
		app, err := c.createApp(ctx)
		if err != nil {
			return nil, err
		}
		appID := to.String(app.AppID)
		appObjID := to.String(app.ObjectID)

		// Create a service principal associated with the new App
		sp, password, err := c.createSP(ctx, app, spExpiration)
		if err != nil {
			c.deleteApp(ctx, appObjID)
			return nil, err
		}

		// Assign Azure roles to the new SP
		raIDs, err := c.assignRoles(ctx, sp, role.AzureRoles)
		if err != nil {
			c.deleteApp(ctx, appObjID)
			return nil, err
		}

		// Assign Azure group memberships to the new SP
		if err := c.addGroupMemberships(ctx, sp, role.AzureGroups); err != nil {
			c.deleteApp(ctx, appObjID)
			return nil, err
		}
	*/
	data := map[string]interface{}{
		"username": role.Username,
		"password": role.Password,
		// "client_id":     appID,
		// "client_secret": password,
	}
	internalData := map[string]interface{}{
		// "app_object_id": appObjID,
		// "sp_object_id":  sp.ObjectID,
		// "role_assignment_ids":  raIDs,
		// "group_membership_ids": groupObjectIDs(role.AzureGroups),
		"role": roleName,
	}

	return b.Secret(SecretTypeSP).Response(data, internalData), nil

}

// createStaticSPSecret adds a new password to the App associated with the role.
func (b *vsphereSecretBackend) createStaticSPSecret(ctx context.Context, c *client, roleName string, role *roleEntry) (*logical.Response, error) {
	lock := locksutil.LockForKey(b.appLocks, role.Username) // We probably need some ID instead of the name  role.ApplicationObjectID)
	lock.Lock()
	defer lock.Unlock()

	govmomiClient, err := c.provider.Login(ctx, role.Username, role.Password, nil)
	if err != nil {
		return nil, err
	}

	marshaledClient, err := govmomiClient.MarshalJSON()
	if err != nil {
		return nil, err
	}

	fmt.Println("marshaledClient=", string(marshaledClient))

	var clientAsMap map[string]interface{}
	err = json.Unmarshal(marshaledClient, &clientAsMap)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"govmomiclient": clientAsMap,
	}
	// TODO: data["cookie"] = the-cookie (?)

	internalData := map[string]interface{}{
		// "app_object_id": role.ApplicationObjectID,
		// "key_id":        keyID,
		"role": roleName,
	}

	return b.Secret(SecretTypeStaticSP).Response(data, internalData), nil
}

func (b *vsphereSecretBackend) spRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleRaw, ok := req.Secret.InternalData["role"]
	if !ok {
		return nil, errors.New("internal data 'role' not found")
	}

	role, err := getRole(ctx, roleRaw.(string), req.Storage)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, nil
	}

	resp := &logical.Response{Secret: req.Secret}
	resp.Secret.TTL = role.TTL
	resp.Secret.MaxTTL = role.MaxTTL

	return resp, nil
}

func (b *vsphereSecretBackend) spRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	resp := new(logical.Response)

	appObjectIDRaw, ok := req.Secret.InternalData["app_object_id"]
	if !ok {
		return nil, errors.New("internal data 'app_object_id' not found")
	}

	appObjectID := appObjectIDRaw.(string)
	fmt.Println("TODO: appObjectID", appObjectID)

	// Get the service principal object ID. Only set if using dynamic service
	// principals.
	var spObjectID string
	if spObjectIDRaw, ok := req.Secret.InternalData["sp_object_id"]; ok {
		spObjectID = spObjectIDRaw.(string)
	}

	var raIDs []string
	if req.Secret.InternalData["role_assignment_ids"] != nil {
		for _, v := range req.Secret.InternalData["role_assignment_ids"].([]interface{}) {
			raIDs = append(raIDs, v.(string))
		}
	}

	var gmIDs []string
	if req.Secret.InternalData["group_membership_ids"] != nil {
		for _, v := range req.Secret.InternalData["group_membership_ids"].([]interface{}) {
			gmIDs = append(gmIDs, v.(string))
		}
	}

	if len(gmIDs) != 0 && spObjectID == "" {
		return nil, errors.New("internal data 'sp_object_id' not found")
	}

	_, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return nil, errwrap.Wrapf("error during revoke: {{err}}", err)
	}

	/*
		// unassigning roles is effectively a garbage collection operation. Errors will be noted but won't fail the
		// revocation process. Deleting the app, however, *is* required to consider the secret revoked.
		if err := c.unassignRoles(ctx, raIDs); err != nil {
			resp.AddWarning(err.Error())
		}

		// removing group membership is effectively a garbage collection
		// operation. Errors will be noted but won't fail the revocation process.
		// Deleting the app, however, *is* required to consider the secret revoked.
		if err := c.removeGroupMemberships(ctx, spObjectID, gmIDs); err != nil {
			resp.AddWarning(err.Error())
		}

		err = c.deleteApp(ctx, appObjectID)
	*/
	return resp, err
}

func (b *vsphereSecretBackend) staticSPRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	return b.logoutFromSession(ctx, req, d)
}

func (b *vsphereSecretBackend) logoutFromSession(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	govmomiClient := &govmomi.Client{}

	clientMarshaled, ok := req.Data["govmomiclient"]
	if !ok {
		return nil, errors.New("data 'govmomiclient' not found")
	}

	clientMarshaledRaw, err := json.Marshal(clientMarshaled)
	if err != nil {
		return nil, err
	}

	// the logged in client:
	err = govmomiClient.UnmarshalJSON(clientMarshaledRaw)
	if err != nil {
		return nil, err
	}

	// TODO: lock this particular session by its cookie?
	// lock := locksutil.LockForKey(b.appLocks, appObjectID)
	// lock.Lock()
	// defer lock.Unlock()

	err = govmomiClient.Logout(ctx)
	if err != nil {
		// should it be just a warning? probably ok as this "just" ends up in the logs. we could filter eventually the "already logged out errors"
		return nil, errwrap.Wrapf("error during revoke: {{err}}", err)
	}
	return nil, nil
}

const pathServicePrincipalHelpSyn = `
Request Service Principal credentials for a given Vault role.
`

const pathServicePrincipalHelpDesc = `
This path creates or updates dynamic Service Principal credentials.
The associated role can be configured to create a new App/Service Principal,
or add a new password to an existing App. The Service Principal or password
will be automatically deleted when the lease has expired.
`
