package vspheresecrets

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/vmware/govmomi"
)

const (
	retryTimeout   = 80 * time.Second
	clientLifetime = 30 * time.Minute // copied from azure... is this relevant here?
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

func (settings *clientSettings) makeLoginURL(username, password string) *url.URL {
	vURL, err := url.Parse(settings.URL) // this is validated earlier
	if err != nil {
		panic(err)
	}
	if username != "" {
		vURL.User = url.UserPassword(username, password)
	}
	return vURL
}

func (settings *clientSettings) Userinfo() *url.Userinfo {
	return url.UserPassword(settings.Username, settings.Password)
}

// makeGovmomiClient returns a new govmomi client. If no username is passed, then no authentication takes place as documented in govmomi.NewClient.
func (settings *clientSettings) makeGovmomiClient(ctx context.Context, username, password string) (*govmomi.Client, error) {
	u := settings.makeLoginURL(username, password)
	client, err := govmomi.NewClient(ctx, u, settings.Insecure)
	return client, err
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

	if b.client.Valid() {
		return b.client, nil
	}

	b.lock.RUnlock()
	b.lock.Lock()
	unlockFunc = b.lock.Unlock

	if b.client.Valid() {
		return b.client, nil
	}

	if b.settings == nil {
		config, err := b.getConfig(ctx, s)
		if err != nil {
			return nil, err
		}
		if config == nil {
			config = new(vsphereConfig)
		}

		settings, err := b.getClientSettings(ctx, config)
		if err != nil {
			return nil, err
		}
		b.settings = settings
	}

	p, err := b.getProvider(ctx, b.settings)
	if err != nil {
		return nil, err
	}

	c := &client{
		provider:   p,
		settings:   b.settings,
		expiration: time.Now().Add(clientLifetime),
	}
	b.client = c

	return c, nil
}
