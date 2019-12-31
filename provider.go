package vspheresecrets

import (
	"context"
	"github.com/vmware/govmomi"
)

// VSphereProvider is an interface to access underlying VSphere govmomi client objects and supporting services.
// Where practical the original function signature is preserved. client provides higher
// level operations ato/ VSphereProvider.
type VSphereProvider interface {
	SessionClient
	// RoleAssignmentsClient
	// RoleDefinitionsClient
}

type SessionClient interface {
	RoleExists(ctx context.Context, role string) (bool, error)
	GroupExists(ctx context.Context, group string) (bool, error)
	UserExists(ctx context.Context, username string) (bool, error)
	Login(ctx context.Context, toBeDefined map[string]interface{}) (*govmomi.Client, error)
	StartSession(ctx context.Context, toBeDefined map[string]interface{}) (string, error)
	RenewSession(ctx context.Context, toBeDefined map[string]interface{}) error
	RevokeSession(ctx context.Context, toBeDefined map[string]interface{}) error
}

// provider is a concrete implementation of AzureProvider. In most cases it is a simple passthrough
// to the appropriate client object. But if the response requires processing that is more practical
// at this layer, the response signature may different from the Azure signature.
type provider struct {
	*clientSettings
	*sessionClient
}

type sessionClient struct {
	settings *clientSettings
}

func (sc *sessionClient) Login(ctx context.Context, toBeDefined map[string]interface{}) (*govmomi.Client, error) {
	return nil, nil
}

func (sc *sessionClient) StartSession(ctx context.Context, toBeDefined map[string]interface{}) (string, error) {
	return "", nil
}

func (sc *sessionClient) RenewSession(ctx context.Context, toBeDefined map[string]interface{}) error {
	return nil
}

func (sc *sessionClient) RevokeSession(ctx context.Context, toBeDefined map[string]interface{}) error {
	return nil
}

func (sc *sessionClient) UserExists(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (sc *sessionClient) RoleExists(ctx context.Context, role string) (bool, error) {
	return false, nil
}

func (sc *sessionClient) GroupExists(ctx context.Context, group string) (bool, error) {
	return false, nil
}

// newVSphereProvider creates an vsphereProvider, backed by VSphere client objects for underlying services.
func newVSphereProvider(settings *clientSettings) (VSphereProvider, error) {
	sc := &sessionClient{
		settings: settings,
	}
	p := &provider{
		clientSettings: settings,
		sessionClient:  sc,
	}
	return p, nil
}
