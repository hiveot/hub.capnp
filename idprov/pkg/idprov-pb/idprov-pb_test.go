package idprovpb_test

import (
	"fmt"
	"github.com/wostzone/wost-go/pkg/logging"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	idprovpb "github.com/wostzone/hub/idprov/pkg/idprov-pb"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/testenv"
)

// testing takes place using the test folder on localhost
var homeFolder string
var certsFolder string

var testCerts testenv.TestCerts

// var hostnames = []string{"localhost"}

// TestMain sets the project test folder as the home folder.
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	certsFolder = path.Join(homeFolder, "certs")
	logging.SetLogging("info", "")
	// certsetup.CreateCertificateBundle(hostnames, certsFolder)
	testCerts = testenv.CreateCertBundle()
	testenv.SaveCerts(&testCerts, certsFolder)

	result := m.Run()

	os.Exit(result)
}

func TestStartStopIdProvPB(t *testing.T) {
	idpConfig := &idprovpb.IDProvPBConfig{}

	hubConfig, err := config.LoadAllConfig(nil, homeFolder, idprovpb.PluginID, &idpConfig)
	assert.NoError(t, err)
	idpPB := idprovpb.NewIDProvPB(idpConfig,
		hubConfig.Address,
		uint(hubConfig.MqttPortCert),
		uint(hubConfig.MqttPortWS),
		testCerts.ServerCert,
		testCerts.CaCert,
		testCerts.CaKey)

	err = idpPB.Start()
	assert.NoError(t, err)

	// Both mqtt and idprov server must live on the same address to be able to use the same server cert
	addrPort := fmt.Sprintf("%s:%d", hubConfig.Address, idpConfig.IdpPort)
	idpc := idprovclient.NewIDProvClient("test", addrPort,
		path.Join(certsFolder, "testCert.pem"),
		path.Join(certsFolder, "testKey.pem"),
		path.Join(certsFolder, config.DefaultCaCertFile))
	err = idpc.Start()
	assert.NoError(t, err)

	idpc.Stop()

	require.NoError(t, err)
	idpPB.Stop()
}
