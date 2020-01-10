package vspheresecrets

import (
	"context"
	"fmt"
	"testing"

	"github.com/hmalphettes/vault-plugin-secrets-vsphere/govmomitest"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/sts"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
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
	ctx := context.Background()
	govmomiClient, err := b.settings.makeGovmomiClient(ctx, govmomitest.SimulatorServerSudoerUsername, govmomitest.SimulatorServerSudoerPassword)
	nilErr(t, err)

	testListDatacenters(t, govmomiClient)

	// Reconstruct the client from its JSON representation and make sure we can use it.
	marshaled, err := govmomiClient.MarshalJSON()
	nilErr(t, err)
	fmt.Println(string(marshaled))

	// this works but we lose the session manager
	cloned := new(vim25.Client)
	err = cloned.UnmarshalJSON(marshaled)
	nilErr(t, err)

	testListDatacenters(t, &govmomi.Client{Client: cloned})

	// try to get a SAML token with the original client:
	ctx2 := context.Background() // must use a separate context or we would get the Session from it.
	stsClient, err := sts.NewClient(ctx2, govmomiClient.Client)
	nilErr(t, err)
	req := sts.TokenRequest{
		// Certificate: c.Certificate(),
		Userinfo:    b.settings.Userinfo(),
		Renewable:   true,
		Delegatable: true,
		// ActAs:       cmd.token != "",
		// Token:       cmd.token,
		// Lifetime: cmd.life,
	}
	signer, err := stsClient.Issue(ctx2, req)
	nilErr(t, err)

	header := soap.Header{Security: signer}

	// The simulator is limited with regard to supporting multiple sessions.
	// we test straight against a brand new session manager... as demonstrated by govmomi/sts/client_test.go
	ctx3 := context.Background() // must use a separate context or we would get the Session from it.
	vimClientNotLogged, err := vim25.NewClient(ctx, soap.NewClient(b.settings.makeLoginURL("", ""), true))
	err = session.NewManager(vimClientNotLogged).LoginByToken(vimClientNotLogged.WithHeader(ctx3, header))
	if err != nil {
		t.Fatal(err)
	}

}

func testListDatacenters(tb testing.TB, govmomiClient *govmomi.Client) {
	finder := find.NewFinder(govmomiClient.Client)
	ctx := context.Background()
	dcs, err := finder.DatacenterList(ctx, "*")
	nilErr(tb, err)
	if len(dcs) != 1 {
		tb.Fatal("Expected to be able to list a single datacenter")
	}

}
