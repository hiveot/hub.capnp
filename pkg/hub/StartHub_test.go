package hub_test

import (
	"flag"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/pkg/hub"
)

var homeFolder string

// Use the test folder during testing
// Reset args to prevent 'flag redefined' error
func setup() {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], strings.Split("", " ")...)
	// hubConfig, _ = hubconfig.SetupConfig(homeFolder, pluginID, customConfig)
}

func TestStartHubNoPlugins(t *testing.T) {
	setup()
	err := hub.StartHub(homeFolder, false)
	assert.NoError(t, err)

	time.Sleep(3 * time.Second)
	hub.StopHub()
}

func TestStartHubWithPlugins(t *testing.T) {
	setup()
	err := hub.StartHub(homeFolder, true)
	assert.NoError(t, err)

	time.Sleep(3 * time.Second)
	hub.StopHub()
}

func TestStartHubBadHome(t *testing.T) {
	setup()
	err := hub.StartHub("/notahomefolder", true)
	assert.Error(t, err)

}
