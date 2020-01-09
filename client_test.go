package vspheresecrets

import (
	"context"
	"fmt"
	"testing"

	"github.com/hmalphettes/vault-plugin-secrets-vsphere/govmomitest"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/sts"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"

	// Register vcsim optional endpoints... we still get a 404 when hitting vsim with an sts call.
	_ "github.com/vmware/govmomi/sts/simulator"
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
	fmt.Println(b.settings)
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
	stsClient, err := sts.NewClient(ctx, govmomiClient.Client)
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
	signer, err := stsClient.Issue(ctx, req)
	nilErr(t, err)

	header := soap.Header{Security: signer}

	err = govmomiClient.SessionManager.LoginByToken(govmomiClient.WithHeader(ctx, header))
	if err != nil {
		t.Fatal(err)
	}

	// govmomiClient.Login(context.Background(), url.UserPass(govmomitest.SimulatorServerSudoerUsername, govmomitest.SimulatorServerSudoerPassword))
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
