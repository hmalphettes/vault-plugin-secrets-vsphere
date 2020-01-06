package vspheresecrets

import (
	"context"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/locksutil"
	"github.com/hashicorp/vault/sdk/logical"
)

// backend wraps the backend framework and adds a map for storing key value pairs
type vsphereSecretBackend struct {
	*framework.Backend

	getProvider func(*clientSettings) (VSphereProvider, error)
	client      *client
	settings    *clientSettings
	lock        sync.RWMutex

	// Creating/deleting passwords against a single Application is a PATCH
	// operation that must be locked per Application Object ID.
	appLocks []*locksutil.LockEntry
}

// Factory configures and returns VSphere backends
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

func backend() *vsphereSecretBackend {
	var b = vsphereSecretBackend{}

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"config",
			},
		},
		Paths: framework.PathAppend(
			pathsRole(&b),
			[]*framework.Path{
				pathConfig(&b),
				// pathServicePrincipal(&b),
			},
		),
		Secrets: []*framework.Secret{
			// secretServicePrincipal(&b),
			// secretStaticServicePrincipal(&b),
		},
		BackendType: logical.TypeLogical,
		Invalidate:  b.invalidate,
	}

	b.getProvider = newVSphereProvider

	return &b
}

// reset clears the backend's cached client
// This is used when the configuration changes and a new client should be
// created with the updated settings.
func (b *vsphereSecretBackend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.settings = nil
}

func (b *vsphereSecretBackend) invalidate(ctx context.Context, key string) {
	switch key {
	case "config":
		b.reset()
	}
}

const backendHelp = `
The VSphere secrets backend dynamically generates VSphere session tokens.
The session tokens have a configurable lease and
are automatically revoked at the end of the lease.

After mounting this backend, credentials to manage VSphere resources
must be configured with the "config/" endpoints and policies must be
written using the "roles/" endpoints before any credentials can be
generated.

Roles can be mapped to existing users or let the admin account create temporary users.
`
