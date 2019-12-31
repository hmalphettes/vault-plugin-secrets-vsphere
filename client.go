package vspheresecrets

import (
	"context"
	"errors"
	"os"
	"strings"

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
	settings.Username = firstAvailable(os.Getenv("GOVMOMI_USERNAME"), config.username)
	settings.Password = firstAvailable(os.Getenv("GOVMOMI_PASSWORD"), config.password)
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
