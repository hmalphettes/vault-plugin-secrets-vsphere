package vspheresecrets

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/logical"
)

// clientSettings is used by a client to configure the connections to Azure.
// It is created from a combination of Vault config settings and environment variables.
type clientSettings struct {
	URL       string
	Username  string
	Password  string
	Insecure  bool
	PluginEnv *logical.PluginEnvironment
}

// getClientSettings creates a new clientSettings object.
// Environment variables have higher precedence than stored configuration.
func (b *vsphereSecretBackend) getClientSettings(ctx context.Context, config *vsphereConfig) (*clientSettings, error) {
	firstAvailable := func(opts ...string) string {
		for _, s := range opts {
			if s != "" {
				return s
			}
		}
		return ""
	}

	settings := new(clientSettings)

	settings.URL = firstAvailable(os.Getenv("GOVMOMI_URL"), config.URL)
	if settings.URL == "" {
		return nil, errors.New("url is required")
	}
	settings.Username = firstAvailable(os.Getenv("GOVMOMI_USERNAME"), config.Username)
	settings.Password = firstAvailable(os.Getenv("GOVMOMI_PASSWORD"), config.Password)
	insecureEnv := os.Getenv("GOVMOMI_INSECURE")
	if insecureEnv != "" {
		settings.Insecure = insecureEnv == "1" || strings.ToLower(insecureEnv) == "true"
	}

	pluginEnv, err := b.System().PluginEnv(ctx)
	if err != nil {
		return nil, errwrap.Wrapf("error loading plugin environment: {{err}}", err)
	}
	settings.PluginEnv = pluginEnv

	return settings, nil
}

// client offers higher level vSphere operations that provide a simpler interface
// for handlers. It in turn relies on a Provider interface to access the lower level
// vSphere Client SDK methods.
type client struct {
	provider   VSphereProvider
	settings   *clientSettings
	expiration time.Time
}

// Valid returns whether the client defined and not expired.
func (c *client) Valid() bool {
	return c != nil && time.Now().Before(c.expiration)
}

func (b *vsphereSecretBackend) getClient(ctx context.Context, s logical.Storage) (*client, error) {
	b.lock.RLock()
	unlockFunc := b.lock.RUnlock
	defer func() { unlockFunc() }()
	return nil, nil
}
