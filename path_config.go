package vspheresecrets

import (
	"context"
	"errors"
	"net/url"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configStoragePath = "config"
)

// vsphereConfig contains values to configure vSphere clients and
// defaults for roles. The zero value is useful and results in
// environments variable and system defaults being used.
type vsphereConfig struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
}

func pathConfig(b *vsphereSecretBackend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"url": &framework.FieldSchema{
				Type: framework.TypeString,
				Description: `ESX or vCenter URL.
				This value can also be provided with the GOVMOMI_URL environment variable.`,
			},
			"username": &framework.FieldSchema{
				Type: framework.TypeString,
				Description: `The username to login to ESX or vCenter. This value can also
				be provided with the GOVMOMI_USERNAME environment variable or via the URL.`,
			},
			"password": &framework.FieldSchema{
				Type: framework.TypeString,
				Description: `The password to login to ESX or vCenter. This value can also
				be provided with the GOVMOMI_PASSWORD environment variable or via the URL.`,
			},
			"insecure": &framework.FieldSchema{
				Type: framework.TypeBool,
				Description: `When true, don't verify the server's certificate chain.
				This value can also be provided with the GOVMOMI_INSECURE environment variable.`,
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathConfigRead,
			logical.CreateOperation: b.pathConfigWrite,
			logical.UpdateOperation: b.pathConfigWrite,
			logical.DeleteOperation: b.pathConfigDelete,
		},
		ExistenceCheck:  b.pathConfigExistenceCheck,
		HelpSynopsis:    confHelpSyn,
		HelpDescription: confHelpDesc,
	}
}

func (b *vsphereSecretBackend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	var merr *multierror.Error

	if config == nil {
		if req.Operation == logical.UpdateOperation {
			return nil, errors.New("config not found during update operation")
		}
		config = new(vsphereConfig)
	}

	if urlD, ok := data.GetOk("url"); ok {
		config.URL = urlD.(string)
		if config.URL == "" {
			merr = multierror.Append(merr, errors.New("The 'url' must not be empty"))
		} else {
			if _, err := url.ParseRequestURI(config.URL); err != nil {
				merr = multierror.Append(merr, err)
			}
		}
	}

	if username, ok := data.GetOk("username"); ok {
		config.Username = username.(string)
	}

	if password, ok := data.GetOk("password"); ok {
		config.Password = password.(string)
	}

	if insecure, ok := data.GetOk("insecure"); ok {
		config.Insecure = insecure.(bool)
	}

	if merr.ErrorOrNil() != nil {
		return logical.ErrorResponse(merr.Error()), nil
	}

	err = b.saveConfig(ctx, config, req.Storage)

	return nil, err
}

func (b *vsphereSecretBackend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.getConfig(ctx, req.Storage)

	if err != nil {
		return nil, err
	}

	if config == nil {
		config = new(vsphereConfig)
	}

	resp := &logical.Response{
		Data: map[string]interface{}{
			"url":      config.URL,
			"username": config.Username,
			// "password": config.Password, // dont return the sensitive secret
			"insecure": config.Insecure,
		},
	}
	return resp, nil
}

func (b *vsphereSecretBackend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, configStoragePath)

	if err == nil {
		b.reset()
	}

	return nil, err
}

func (b *vsphereSecretBackend) pathConfigExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	config, err := b.getConfig(ctx, req.Storage)
	if err != nil {
		return false, err
	}

	return config != nil, err
}

func (b *vsphereSecretBackend) getConfig(ctx context.Context, s logical.Storage) (*vsphereConfig, error) {
	entry, err := s.Get(ctx, configStoragePath)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	config := new(vsphereConfig)
	if err := entry.DecodeJSON(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (b *vsphereSecretBackend) saveConfig(ctx context.Context, config *vsphereConfig, s logical.Storage) error {
	entry, err := logical.StorageEntryJSON(configStoragePath, config)

	if err != nil {
		return err
	}

	err = s.Put(ctx, entry)
	if err != nil {
		return err
	}

	// reset the backend since the client and provider will have been
	// built using old versions of this data
	b.reset()

	return nil
}

const confHelpSyn = `Configure the vSphere Secret backend.`
const confHelpDesc = `
The vSphere secret backend requires credentials for managing users.
This endpoint is used to configure those credentials as
well as default values for the backend in general.
`
