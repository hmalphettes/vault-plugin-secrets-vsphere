package vspheresecrets

import (
	"context"

	"github.com/vmware/govmomi"
)

// VSphereProvider is an interface to access underlying VSphere govmomi client objects and supporting services.
// Where practical the original function signature is preserved. client provides higher
// level operations ato/ VSphereProvider.
type VSphereProvider interface {
	GetMountGovmomiClient() *govmomi.Client
	RoleExists(ctx context.Context, role string) (bool, error)
	GroupExists(ctx context.Context, group string) (bool, error)
	UserExists(ctx context.Context, username string) (bool, error)
	Login(ctx context.Context, username, password string, toBeDefined map[string]interface{}) (*govmomi.Client, error)
	StartSession(ctx context.Context, toBeDefined map[string]interface{}) (string, error)
	RenewSession(ctx context.Context, toBeDefined map[string]interface{}) error
	RevokeSession(ctx context.Context, toBeDefined map[string]interface{}) error
}

// provider is a concrete implementation of vSphereProvider. In most cases it is a simple passthrough
// to the appropriate client object. But if the response requires processing that is more practical
// at this layer, the response signature may different from the vSphere signature.
type provider struct {
	settings      *clientSettings
	govmomiClient *govmomi.Client
}

// GetMountGovmomiClient returns the underlying govmami.Client using the credentials defined in the config of the mount.
// When no credentials are defined, the client cannot login and authenticate.
func (p *provider) GetMountGovmomiClient() *govmomi.Client {
	return p.govmomiClient
}

// Login authenticate with the credentials defined on the role
func (p *provider) Login(ctx context.Context, username, password string, toBeDefined map[string]interface{}) (*govmomi.Client, error) {
	return p.settings.makeGovmomiClient(ctx, username, password)
}

func (p *provider) StartSession(ctx context.Context, toBeDefined map[string]interface{}) (string, error) {
	return "", nil
}

func (p *provider) RenewSession(ctx context.Context, toBeDefined map[string]interface{}) error {
	return nil
}

func (p *provider) RevokeSession(ctx context.Context, toBeDefined map[string]interface{}) error {
	return nil
}

func (p *provider) UserExists(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (p *provider) RoleExists(ctx context.Context, role string) (bool, error) {
	return false, nil
}

func (p *provider) GroupExists(ctx context.Context, group string) (bool, error) {
	return false, nil
}

// newVSphereProvider creates an vsphereProvider, backed by VSphere client objects for underlying services.
func newVSphereProvider(ctx context.Context, settings *clientSettings) (VSphereProvider, error) {
	govmomiClient, err := settings.makeGovmomiClient(ctx, settings.Username, settings.Password)
	if err != nil {
		return nil, err
	}

	p := &provider{
		govmomiClient: govmomiClient,
		settings:      settings,
	}
	return p, nil
}
