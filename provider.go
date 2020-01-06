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
	Login(ctx context.Context, role string, toBeDefined map[string]interface{}) (*govmomi.Client, error)
	StartSession(ctx context.Context, toBeDefined map[string]interface{}) (string, error)
	RenewSession(ctx context.Context, toBeDefined map[string]interface{}) error
	RevokeSession(ctx context.Context, toBeDefined map[string]interface{}) error
}

// provider is a concrete implementation of vSphereProvider. In most cases it is a simple passthrough
// to the appropriate client object. But if the response requires processing that is more practical
// at this layer, the response signature may different from the vSphere signature.
type provider struct {
	govmomiClient *govmomi.Client
}

// GetMountGovmomiClient returns the underlying govmami.Client using the credentials defined in the config of the mount.
// When no credentials are defined, the client cannot login and authenticate.
func (p *provider) GetMountGovmomiClient() *govmomi.Client {
	return p.govmomiClient
}

func (sc *provider) Login(ctx context.Context, role string, toBeDefined map[string]interface{}) (*govmomi.Client, error) {
	return nil, nil
}

func (sc *provider) StartSession(ctx context.Context, toBeDefined map[string]interface{}) (string, error) {
	return "", nil
}

func (sc *provider) RenewSession(ctx context.Context, toBeDefined map[string]interface{}) error {
	return nil
}

func (sc *provider) RevokeSession(ctx context.Context, toBeDefined map[string]interface{}) error {
	return nil
}

func (sc *provider) UserExists(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (sc *provider) RoleExists(ctx context.Context, role string) (bool, error) {
	return false, nil
}

func (sc *provider) GroupExists(ctx context.Context, group string) (bool, error) {
	return false, nil
}

// newVSphereProvider creates an vsphereProvider, backed by VSphere client objects for underlying services.
func newVSphereProvider(settings *clientSettings) (VSphereProvider, error) {
	govmomiClient, err := settings.makeGovmomiClient(settings.Username, settings.Password)
	if err != nil {
		return nil, err
	}

	p := &provider{
		govmomiClient: govmomiClient,
	}
	return p, nil
}
