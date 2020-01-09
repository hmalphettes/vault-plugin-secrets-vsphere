package govmomitest

import (
	"net/url"
	"testing"
	"time"

	"github.com/vmware/govmomi/simulator"

	// Register vcsim optional endpoints
	_ "github.com/vmware/govmomi/lookup/simulator"
	_ "github.com/vmware/govmomi/pbm/simulator"
	_ "github.com/vmware/govmomi/sts/simulator"
	_ "github.com/vmware/govmomi/vapi/simulator"
)

const (
	// SimulatorServerSudoerUsername Username set on the VPX simulator model.
	SimulatorServerSudoerUsername = "testvaultsudoer"
	// SimulatorServerSudoerPassword password set on the VPX simulator model.
	SimulatorServerSudoerPassword = "sudo"
	defaultLeaseTTLHr             = 1 * time.Hour
	maxLeaseTTLHr                 = 12 * time.Hour
	defaultTestTTL                = 300
	defaultTestMaxTTL             = 3600
)

var model *simulator.Model
var server *simulator.Server

// SimulatorURL URL of the vSphere endpoint backed by the simulator
var SimulatorURL string

// SimulatorConfig Config parameters for the vSphere secrets that match the vSphere endpoint backed by the simulator
var simulatorConfig map[string]interface{}

// CreateSimulator sets up a govmomi simulator model configured with a user
func createSimulator(t *testing.T) *simulator.Model {
	// Default vCenter model. We may end-up customizing this.
	model := simulator.VPX()

	// defer model.Remove()
	err := model.Create()
	if err != nil {
		t.Error(err)
		return nil
	}

	model.Service.Listen = &url.URL{
		User: url.UserPassword(SimulatorServerSudoerUsername, SimulatorServerSudoerPassword),
	}

	return model
}

// Setup creates a simulator if not in place
func Setup(t *testing.T) *simulator.Server {
	if server != nil {
		return server
	}
	model = createSimulator(t)
	if model != nil {
		server = model.Service.NewServer()
		if server != nil {
			SimulatorURL = server.URL.String()
		}
		simulatorConfig = map[string]interface{}{
			"url":      SimulatorURL,
			"username": SimulatorServerSudoerUsername,
			"password": SimulatorServerSudoerPassword,
			"insecure": true,
			// "ttl":      defaultTestTTL,
			// "max_ttl":  defaultTestMaxTTL,
		}
	}
	return server
}

// TearDown cleans up the simulator
func TearDown() {
	if server != nil {
		server.Close()
	}
	if model != nil {
		model.Remove()
	}
	simulatorConfig = nil
	SimulatorURL = ""
}

func GetSimulatorConfig(withPassword bool) map[string]interface{} {
	cfg := make(map[string]interface{})
	for k, v := range simulatorConfig {
		if withPassword || k != "password" {
			cfg[k] = v
		}
	}
	return cfg
}
