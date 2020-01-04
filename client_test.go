package vspheresecrets

import (
	"testing"

	"github.com/hmalphettes/vault-plugin-secrets-vsphere/govmomitest"
)

func TestProvider(t *testing.T) {
	server := govmomitest.Setup(t)
	defer govmomitest.TearDown()

	url := server.URL
	t.Log(url)
}
