package vspheresecrets

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/sdk/logical"
)

var (
	testStaticSPRole = map[string]interface{}{
		"username": "foo",
		"password": "bar",
	}
)

func TestStaticSPRead(t *testing.T) {
	b, s := getTestBackend(t, true)

	// verify basic cred issuance
	t.Run("Basic", func(t *testing.T) {
		name := generateUUID()
		testRoleCreate(t, b, s, name, testStaticSPRole)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "creds/" + name,
			Storage:   s,
		})

		nilErr(t, err)

		if resp.IsError() {
			t.Fatalf("expected no response error, actual:%#v", resp.Error())
		}

		// verify client_id format, and that the corresponding app actually exists
		_, err = uuid.ParseUUID(resp.Data["client_id"].(string))
		nilErr(t, err)

		keyID := resp.Secret.InternalData["key_id"].(string)
		if !strings.HasPrefix(keyID, "ffffff") {
			t.Fatalf("expected prefix 'ffffff': %s", keyID)
		}

		// client, err := b.getClient(context.Background(), s)
		// nilErr(t, err)

		// if !client.provider.(*mockProvider).passwordExists(keyID) {
		// 	t.Fatalf("password was not created")
		// }

		// verify password format
		_, err = uuid.ParseUUID(resp.Data["client_secret"].(string))
		nilErr(t, err)
	})

	// verify role TTLs are reflected in secret
	t.Run("TTLs", func(t *testing.T) {
		name := generateUUID()
		testRoleCreate(t, b, s, name, testStaticSPRole)

		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "session/" + name,
			Storage:   s,
		})

		nilErr(t, err)

		equal(t, 0*time.Second, resp.Secret.TTL)
		equal(t, 0*time.Second, resp.Secret.MaxTTL)

		roleUpdate := map[string]interface{}{
			"ttl":     20,
			"max_ttl": 30,
		}
		testRoleCreate(t, b, s, name, roleUpdate)

		resp, err = b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "session/" + name,
			Storage:   s,
		})

		nilErr(t, err)

		equal(t, 20*time.Second, resp.Secret.TTL)
		equal(t, 30*time.Second, resp.Secret.MaxTTL)
	})
}

func TestStaticSPRevoke(t *testing.T) {
	b, s := getTestBackend(t, true)

	testRoleCreate(t, b, s, "test_role", testStaticSPRole)

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "session/test_role",
		Storage:   s,
	})

	keyID := resp.Secret.InternalData["key_id"].(string)
	if !strings.HasPrefix(keyID, "ffffff") {
		t.Fatalf("expected prefix 'ffffff': %s", keyID)
	}

	// client, err := b.getClient(context.Background(), s)
	// nilErr(t, err)

	// if !client.provider.(*mockProvider).passwordExists(keyID) {
	// 	t.Fatalf("password was not created")
	// }

	// Serialize and deserialize the secret to remove typing, as will really happen.
	fakeSaveLoad(resp.Secret)

	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.RevokeOperation,
		Secret:    resp.Secret,
		Storage:   s,
	})

	nilErr(t, err)

	if resp.IsError() {
		t.Fatalf("receive response error: %v", resp.Error())
	}

	// if client.provider.(*mockProvider).passwordExists(keyID) {
	// 	t.Fatalf("password present but should have been deleted")
	// }
}
