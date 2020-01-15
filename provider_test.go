package vspheresecrets

import (
	"context"
	"testing"
	"time"

	"github.com/hmalphettes/vault-plugin-secrets-vsphere/govmomitest"
)

func TestProvider(t *testing.T) {
	server := govmomitest.Setup(t)
	defer govmomitest.TearDown()

	url := server.URL
	t.Log(url)

	b, storage := getTestBackend(t, true)
	if storage == nil {
		t.Fatal("The storage must not be nil")
	}
	provider, err := b.getProvider(context.Background(), b.settings)
	nilErr(t, err)
	t.Run("Test provider.Login", func(t *testing.T) {
		ctx := context.Background()
		c, err := provider.Login(ctx, govmomitest.SimulatorServerSudoerUsername, govmomitest.SimulatorServerSudoerPassword, nil)
		nilErr(t, err)
		testListDatacenters(t, c)
	})
	t.Run("Test provider.IssueUserToken", func(t *testing.T) {
		ctx := context.Background()
		signer, err := provider.IssueUserToken(ctx, govmomitest.SimulatorServerSudoerUsername, govmomitest.SimulatorServerSudoerPassword, 30*time.Second, true, true)
		nilErr(t, err)
		c, err := makeGovmomiClientFromToken(ctx, b, signer)
		nilErr(t, err)
		testListDatacenters(t, c)
	})

	t.Run("Test provider.IssueSolutionToken", func(t *testing.T) {
		t.Skip("Not supported yet")
	})
}
