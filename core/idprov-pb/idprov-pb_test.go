package idprovpb_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	idprovpb "github.com/wostzone/hub/core/idprov-pb"
	"github.com/wostzone/idprov-go/pkg/idprov"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// testing takes place using the test folder on localhost
var homeFolder string
var certsFolder string

var hostnames = []string{"localhost"}

// TestMain sets the project test folder as the home folder.
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	certsFolder = path.Join(homeFolder, "certs")
	certsetup.CreateCertificateBundle(hostnames, certsFolder)

	result := m.Run()

	os.Exit(result)
}

func TestStartStopIdProvPB(t *testing.T) {
	idpConfig := &idprovpb.IDProvPBConfig{}

	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, idprovpb.PluginID, &idpConfig)
	assert.NoError(t, err)
	idpPB := idprovpb.NewIDProvPB(idpConfig, hubConfig)

	err = idpPB.Start()
	assert.NoError(t, err)

	// Both mqtt and idprov server must live on the same address to be able to use the same server cert
	idpc := idprov.NewIDProvClient("test", hubConfig.MqttAddress, idpConfig.IdpPort, hubConfig.CertsFolder)
	err = idpc.Start()
	assert.NoError(t, err)

	idpc.Stop()

	require.NoError(t, err)
	idpPB.Stop()
}
