package vspheresecrets

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	rolesStoragePath = "roles"

	credentialTypeSP = 0
)

// roleEntry is a Vault role construct that maps to Azure roles or Applications
type roleEntry struct {
	// CredentialType      int           `json:"credential_type"` // Reserved. Always SP at this time.
	// AzureRoles          []*AzureRole  `json:"azure_roles"`
	// AzureGroups         []*AzureGroup `json:"azure_groups"`
	Username string `json:"username"`
	Password string `json:"password"` // make sure we dont serialize it back though
	// ApplicationObjectID string        `json:"application_object_id"`
	TTL           time.Duration `json:"ttl"`
	VSphereRoles  []string      `json:"vsphere_roles"`
	VSphereGroups []string      `json:"vsphere_groups"`
	MaxTTL        time.Duration `json:"max_ttl"`
}

func pathsRole(b *vsphereSecretBackend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "roles/" + framework.GenericNameRegex("name"),
			Fields: map[string]*framework.FieldSchema{
				"name": {
					Type:        framework.TypeLowerCaseString,
					Description: "Name of the role.",
				},
				"username": {
					Type:        framework.TypeString,
					Description: "Optional username to use. Or existing username (when password is defined). Each '?' character is replaced by a random a-z0-9 character for each call. When empty, the default value is vault-{role}-???",
				},
				"password": {
					Type:        framework.TypeString,
					Description: "Optional password to use. When defined, no users are created.",
				},
				"vsphere_roles": {
					Type:        framework.TypeCommaStringSlice,
					Description: "Comma separated list of VSphere roles to assign - when password is empty.",
				},
				"vsphere_groups": {
					Type:        framework.TypeCommaStringSlice,
					Description: "Comma separated list of VSphere groups to assign the temporary user to - when password is empty.",
				},
				"ttl": {
					Type:        framework.TypeDurationSecond,
					Description: "Default lease for generated credentials. If not set or set to 0, will use system default.",
				},
				"max_ttl": {
					Type:        framework.TypeDurationSecond,
					Description: "Maximum time a service principal. If not set or set to 0, will use system default.",
				},
			},
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ReadOperation:   b.pathRoleRead,
				logical.CreateOperation: b.pathRoleUpdate,
				logical.UpdateOperation: b.pathRoleUpdate,
				logical.DeleteOperation: b.pathRoleDelete,
			},
			HelpSynopsis:    roleHelpSyn,
			HelpDescription: roleHelpDesc,
			ExistenceCheck:  b.pathRoleExistenceCheck,
		},
		{
			Pattern: "roles/?",
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: b.pathRoleList,
			},
			HelpSynopsis:    roleListHelpSyn,
			HelpDescription: roleListHelpDesc,
		},
	}

}

// pathRoleUpdate creates or updates Vault roles.
//
// Basic validity check are made to verify that the provided fields meet requirements
// for the given credential type.
//
// Dynamic Service Principal:
//   Azure roles are checked for existence. The Azure role lookup step will allow the
//   operator to provide a role name or ID. ID is unambigious and will be used if provided.
//   Given just role name, a search will be performed and if exactly one match is found,
//   that role will be used.

//   Azure groups are checked for existence. The Azure groups lookup step will allow the
//   operator to provide a groups name or ID. ID is unambigious and will be used if provided.
//   Given just group name, a search will be performed and if exactly one match is found,
//   that group will be used.
//
// Static Service Principal:
//   The provided Application Object ID is checked for existence.
func (b *vsphereSecretBackend) pathRoleUpdate(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	var resp *logical.Response

	// client, err := b.getClient(ctx, req.Storage)
	// if err != nil {
	// 	return nil, err
	// }

	// load or create role
	name := d.Get("name").(string)
	role, err := getRole(ctx, name, req.Storage)
	if err != nil {
		return nil, errwrap.Wrapf("error reading role: {{err}}", err)
	}

	if role == nil {
		if req.Operation == logical.UpdateOperation {
			return nil, errors.New("role entry not found during update operation")
		}
		role = &roleEntry{}
	}

	// load and validate TTLs
	if ttlRaw, ok := d.GetOk("ttl"); ok {
		role.TTL = time.Duration(ttlRaw.(int)) * time.Second
	} else if req.Operation == logical.CreateOperation {
		role.TTL = time.Duration(d.Get("ttl").(int)) * time.Second
	}

	if maxTTLRaw, ok := d.GetOk("max_ttl"); ok {
		role.MaxTTL = time.Duration(maxTTLRaw.(int)) * time.Second
	} else if req.Operation == logical.CreateOperation {
		role.MaxTTL = time.Duration(d.Get("max_ttl").(int)) * time.Second
	}

	if role.MaxTTL != 0 && role.TTL > role.MaxTTL {
		return logical.ErrorResponse("ttl cannot be greater than max_ttl"), nil
	}

	// update and verify Application Object ID if provided
	if username, ok := d.GetOk("username"); ok {
		role.Username = username.(string)
	}

	if role.Username != "" && role.Password != "" {
		fmt.Println("Username and password pre-set")
		// TODO: check for the user to be defined already
		// app, err := client.provider.GetApplication(ctx, role.Username)
		// if err != nil {
		// 	return nil, errwrap.Wrapf("error loading Application: {{err}}", err)
		// }
		// role.ApplicationID = to.String(app.AppID)
	} else if role.Username == "" {
		role.Username = name + "-???"
	}

	// Parse the VSPhere roles
	if roles, ok := d.GetOk("vsphere_roles"); ok {
		role.VSphereRoles = roles.([]string)
	}

	// Parse the Azure groups
	if groups, ok := d.GetOk("vsphere_groups"); ok {
		role.VSphereGroups = groups.([]string)
	}

	// TODO for VSphere:
	// verify VSPhere roles, including looking up each role by ID or name.
	// roleSet := make(map[string]bool)
	// for _, r := range role.VSphereRoles {
	// 	var roleDef authorization.RoleDefinition
	// 	if r.RoleID != "" {
	// 		roleDef, err = client.provider.GetRoleByID(ctx, r.RoleID)
	// 		if err != nil {
	// 			if strings.Contains(err.Error(), "RoleDefinitionDoesNotExist") {
	// 				return logical.ErrorResponse("no role found for role_id: '%s'", r.RoleID), nil
	// 			}
	// 			return nil, errwrap.Wrapf("unable to lookup Azure role: {{err}}", err)
	// 		}
	// 	} else {
	// 		defs, err := client.findRoles(ctx, r.RoleName)
	// 		if err != nil {
	// 			return nil, errwrap.Wrapf("unable to lookup Azure role: {{err}}", err)
	// 		}
	// 		if l := len(defs); l == 0 {
	// 			return logical.ErrorResponse("no role found for role_name: '%s'", r.RoleName), nil
	// 		} else if l > 1 {
	// 			return logical.ErrorResponse("multiple matches found for role_name: '%s'. Specify role by ID instead.", r.RoleName), nil
	// 		}
	// 		roleDef = defs[0]
	// 	}

	// 	roleDefID := to.String(roleDef.ID)
	// 	roleDefName := to.String(roleDef.RoleName)

	// 	r.RoleName, r.RoleID = roleDefName, roleDefID

	// 	rsKey := r.RoleID + "||" + r.Scope
	// 	if roleSet[rsKey] {
	// 		return logical.ErrorResponse("duplicate role_id and scope: '%s', '%s'", r.RoleID, r.Scope), nil
	// 	}
	// 	roleSet[rsKey] = true
	// }

	// TODO for VSphere:
	// update and verify Azure groups, including looking up each group by ID or name.
	// groupSet := make(map[string]bool)
	// for _, r := range role.AzureGroups {
	// 	var groupDef graphrbac.ADGroup
	// 	if r.ObjectID != "" {
	// 		groupDef, err = client.provider.GetGroup(ctx, r.ObjectID)
	// 		if err != nil {
	// 			if strings.Contains(err.Error(), "Request_ResourceNotFound") {
	// 				return logical.ErrorResponse("no group found for object_id: '%s'", r.ObjectID), nil
	// 			}
	// 			return nil, errwrap.Wrapf("unable to lookup Azure group: {{err}}", err)
	// 		}
	// 	} else {
	// 		defs, err := client.findGroups(ctx, r.GroupName)
	// 		if err != nil {
	// 			return nil, errwrap.Wrapf("unable to lookup Azure group: {{err}}", err)
	// 		}
	// 		if l := len(defs); l == 0 {
	// 			return logical.ErrorResponse("no group found for group_name: '%s'", r.GroupName), nil
	// 		} else if l > 1 {
	// 			return logical.ErrorResponse("multiple matches found for group_name: '%s'. Specify group by ObjectID instead.", r.GroupName), nil
	// 		}
	// 		groupDef = defs[0]
	// 	}

	// 	groupDefID := to.String(groupDef.ObjectID)
	// 	groupDefName := to.String(groupDef.DisplayName)
	// 	r.GroupName, r.ObjectID = groupDefName, groupDefID

	// 	if groupSet[r.ObjectID] {
	// 		return logical.ErrorResponse("duplicate object_id '%s'", r.ObjectID), nil
	// 	}
	// 	groupSet[r.ObjectID] = true
	// }

	if role.Password == "" && len(role.VSphereRoles) == 0 && len(role.VSphereGroups) == 0 {
		return logical.ErrorResponse("either VSphere role definitions, group definitions, or a username and password must be provided"), nil
	}

	// save role
	err = saveRole(ctx, req.Storage, role, name)
	if err != nil {
		return nil, errwrap.Wrapf("error storing role: {{err}}", err)
	}

	return resp, nil
}

func (b *vsphereSecretBackend) pathRoleRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	var data = make(map[string]interface{})

	name := d.Get("name").(string)

	r, err := getRole(ctx, name, req.Storage)
	if err != nil {
		return nil, errwrap.Wrapf("error reading role: {{err}}", err)
	}

	if r == nil {
		return nil, nil
	}

	data["ttl"] = r.TTL / time.Second
	data["max_ttl"] = r.MaxTTL / time.Second
	data["vsphere_roles"] = r.VSphereRoles
	data["vsphere_groups"] = r.VSphereGroups
	data["username"] = r.Username
	data["password"] = r.Password

	return &logical.Response{
		Data: data,
	}, nil
}

func (b *vsphereSecretBackend) pathRoleList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roles, err := req.Storage.List(ctx, rolesStoragePath+"/")
	if err != nil {
		return nil, errwrap.Wrapf("error listing roles: {{err}}", err)
	}

	return logical.ListResponse(roles), nil
}

func (b *vsphereSecretBackend) pathRoleDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)

	err := req.Storage.Delete(ctx, fmt.Sprintf("%s/%s", rolesStoragePath, name))
	if err != nil {
		return nil, errwrap.Wrapf("error deleting role: {{err}}", err)
	}

	return nil, nil
}

func (b *vsphereSecretBackend) pathRoleExistenceCheck(ctx context.Context, req *logical.Request, d *framework.FieldData) (bool, error) {
	name := d.Get("name").(string)

	role, err := getRole(ctx, name, req.Storage)
	if err != nil {
		return false, errwrap.Wrapf("error reading role: {{err}}", err)
	}

	return role != nil, nil
}

func saveRole(ctx context.Context, s logical.Storage, c *roleEntry, name string) error {
	entry, err := logical.StorageEntryJSON(fmt.Sprintf("%s/%s", rolesStoragePath, name), c)
	if err != nil {
		return err
	}

	return s.Put(ctx, entry)
}

func getRole(ctx context.Context, name string, s logical.Storage) (*roleEntry, error) {
	entry, err := s.Get(ctx, fmt.Sprintf("%s/%s", rolesStoragePath, name))
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	role := new(roleEntry)
	if err := entry.DecodeJSON(role); err != nil {
		return nil, err
	}
	return role, nil
}

const roleHelpSyn = "Manage the Vault roles used to generate VSphere credentials."
const roleHelpDesc = `
This path allows you to read and write roles that are used to generate VSphere login
credentials. These roles are associated with either an existing user, or a set
of VSphere roles and groups, which are used to control permissions to VSphere resources.

If the backend is mounted at "vsphere", you would create a Vault role at "vsphere/roles/my_role",
and request credentials from "azure/creds/my_role".

Each Vault role is configured with the standard ttl parameters and either an
username/password or a combination of VSphere roles and groups to make the dynamically created
user a member of, and VSphere roles to assign the dynamically created
user to. During the Vault role creation, any set VSphere role, group, or
Object ID will be fetched and verified, and therefore must exist for the request
to succeed. When a user requests credentials against the Vault role, a new
user will be created if the password field is empty. In that case the new user is assigned
its roles and added to the groups. Then the user is logged in and the session-token is returned by vault.
Otherwise, the existing username/password is submitted to VSPhere to retrieve a new session-token and returned by vault.
`
const roleListHelpSyn = `List existing roles.`
const roleListHelpDesc = `List existing roles by name.`
