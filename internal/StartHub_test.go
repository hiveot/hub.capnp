package internal_test

import (
	"flag"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/internal"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
)

// testing takes place using the test folder on localhost
var homeFolder string
var certsFolder string

const hostname = "localhost"

// Use the test folder during testing
// Reset args to prevent 'flag redefined' error
func setup() {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")
	certsFolder = path.Join(homeFolder, "certs")
	certsetup.CreateCertificateBundle(hostname, certsFolder)

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = append(os.Args[0:1], strings.Split("", " ")...)
	// hubConfig, _ = hubconfig.SetupConfig(homeFolder, pluginID, customConfig)

}

func TestStartHubNoPlugins(t *testing.T) {
	setup()
	err := internal.StartHub(homeFolder, false)
	require.NoError(t, err)

	time.Sleep(3 * time.Second)
	internal.StopHub()
}

func TestStartHubWithPlugins(t *testing.T) {
	setup()
	err := internal.StartHub(homeFolder, true)
	assert.NoError(t, err)

	time.Sleep(3 * time.Second)
	internal.StopHub()
}

func TestStartHubBadHome(t *testing.T) {
	setup()
	err := internal.StartHub("/notahomefolder", true)
	assert.Error(t, err)

}
