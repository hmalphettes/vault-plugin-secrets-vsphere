package vspheresecrets

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hmalphettes/vault-plugin-secrets-vsphere/govmomitest"
)

func TestConfig(t *testing.T) {
	_ = govmomitest.Setup(t)
	defer govmomitest.TearDown()
	b, s := getTestBackend(t, false)

	// Test valid config
	config := govmomitest.GetSimulatorConfig(true)

	testConfigCreate(t, b, s, config)

	// Must not be able to retrieve the password from the read of a config
	delete(config, "password")
	testConfigRead(t, b, s, config)

	// Test test updating one element retains the others
	config["username"] = "different"
	configSubset := map[string]interface{}{
		"username": config["username"],
	}
	testConfigCreate(t, b, s, configSubset)
	testConfigUpdate(t, b, s, config)

	// Test bad environment
	config = map[string]interface{}{
		"url": "invalidURL",
	}

	resp, _ := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data:      config,
		Storage:   s,
	})

	if !resp.IsError() {
		t.Fatal("expected a response error")
	}
}

func TestConfigDelete(t *testing.T) {
	_ = govmomitest.Setup(t)
	defer govmomitest.TearDown()

	b, s := getTestBackend(t, false)

	// Test valid config
	config := govmomitest.GetSimulatorConfig(true)

	testConfigCreate(t, b, s, config)

	delete(config, "password")
	testConfigRead(t, b, s, config)

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "config",
		Storage:   s,
	})

	nilErr(t, err)

	if resp != nil && resp.IsError() {
		t.Fatal(resp.Error())
	}

	config = map[string]interface{}{
		"url":      "",
		"username": "",
		"insecure": false,
	}
	testConfigRead(t, b, s, config)
}

func testConfigCreate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]interface{}) {
	t.Helper()
	testConfigCreateUpdate(t, b, logical.CreateOperation, s, d)
}

func testConfigUpdate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]interface{}) {
	t.Helper()
	testConfigCreateUpdate(t, b, logical.UpdateOperation, s, d)
}

func testConfigCreateUpdate(t *testing.T, b logical.Backend, op logical.Operation, s logical.Storage, d map[string]interface{}) {
	t.Helper()

	// save and restore the client since the config change will clear it
	settings := b.(*vsphereSecretBackend).settings
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: op,
		Path:      "config",
		Data:      d,
		Storage:   s,
	})
	b.(*vsphereSecretBackend).settings = settings

	if err != nil {
		t.Fatal(err)
	}

	if resp != nil && resp.IsError() {
		t.Fatal(resp.Error())
	}
}

func testConfigRead(t *testing.T, b logical.Backend, s logical.Storage, expected map[string]interface{}) {
	t.Helper()
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "config",
		Storage:   s,
	})

	if err != nil {
		t.Fatal(err)
	}

	if resp != nil && resp.IsError() {
		t.Fatal(resp.Error())
	}

	equal(t, expected, resp.Data)
}
