package idprovpb_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	idprovpb "github.com/wostzone/hub/core/idprov-pb"
	"github.com/wostzone/idprov-go/pkg/idprov"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

var homeFolder string

// TestMain sets the project test folder as the home folder.
// Make sure the certificates exist.
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	result := m.Run()

	os.Exit(result)
}

func TestStartStop(t *testing.T) {
	idpConfig := &idprovpb.IDProvPBConfig{}

	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, idprovpb.PluginID, &idpConfig)
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
