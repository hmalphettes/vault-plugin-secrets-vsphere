package vspheresecrets

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hmalphettes/vault-plugin-secrets-vsphere/govmomitest"

	log "github.com/hashicorp/go-hclog"
)

const (
	defaultLeaseTTLHr = 1 * time.Hour
	maxLeaseTTLHr     = 12 * time.Hour
	defaultTestTTL    = 300
	defaultTestMaxTTL = 3600
)

func getTestBackend(t *testing.T, initConfig bool) (*vsphereSecretBackend, logical.Storage) {
	b := backend()

	config := &logical.BackendConfig{
		Logger: logging.NewVaultLogger(log.Trace),
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: defaultLeaseTTLHr,
			MaxLeaseTTLVal:     maxLeaseTTLHr,
		},
		StorageView: &logical.InmemStorage{},
	}
	err := b.Setup(context.Background(), config)
	if err != nil {
		t.Fatalf("unable to create backend: %v", err)
	}

	b.settings = &clientSettings{
		URL:      govmomitest.SimulatorURL,
		Username: govmomitest.SimulatorServerSudoerUsername,
		Password: govmomitest.SimulatorServerSudoerPassword,
		Insecure: true,
	}
	// mockProvider := newMockProvider()
	// b.getProvider = func(s *clientSettings) (VSphereProvider, error) {
	// 	return mockProvider, nil
	// }

	if initConfig {
		cfg := govmomitest.GetSimulatorConfig(true)
		cfg["ttl"] = defaultTestTTL
		cfg["max_ttl"] = defaultTestMaxTTL

		testConfigCreate(t, b, config.StorageView, cfg)
	}

	return b, config.StorageView
}
