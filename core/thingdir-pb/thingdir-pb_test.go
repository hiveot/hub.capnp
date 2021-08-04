package thingdirpb_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	thingdirpb "github.com/wostzone/hub/core/thingdir-pb"
	"github.com/wostzone/thingdir-go/pkg/dirclient"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
	"github.com/wostzone/wostlib-go/pkg/td"
	"github.com/wostzone/wostlib-go/pkg/testenv"
	"github.com/wostzone/wostlib-go/pkg/vocab"
)

var certFolder string

// var serverAddress string

var homeFolder string
var hubConfig *hubconfig.HubConfig

var caCertPath string
var serverCertPath string
var serverKeyPath string
var pluginCertPath string
var pluginKeyPath string

// TestMain runs a directory server for use by the test cases in this package
// This uses the directory client in testing
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of DirectoryServer ------")
	// serverAddress = hubnet.GetOutboundIP("").String()

	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	certFolder = path.Join(homeFolder, "certs")
	hubConfig, _ = hubconfig.LoadHubConfig(homeFolder, "plugin1")
	hostnames := []string{hubConfig.MqttAddress}

	// make sure the certificates are there
	certsetup.CreateCertificateBundle(hostnames, certFolder)
	serverCertPath = path.Join(certFolder, certsetup.HubCertFile)
	serverKeyPath = path.Join(certFolder, certsetup.HubKeyFile)
	caCertPath = path.Join(certFolder, certsetup.CaCertFile)
	pluginCertPath = path.Join(certFolder, certsetup.PluginCertFile)
	pluginKeyPath = path.Join(certFolder, certsetup.PluginKeyFile)

	// start a mqtt test server
	mosquittoCmd := testenv.Setup(homeFolder, hubConfig.MqttCertPort)

	res := m.Run()

	testenv.Teardown(mosquittoCmd)

	os.Exit(res)
}

func TestStartStopThingDirectoryService(t *testing.T) {
	tdirConfig := &thingdirpb.ThingDirPBConfig{DirAddress: hubConfig.MqttAddress}
	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, thingdirpb.PluginID, &tdirConfig)
	assert.NoError(t, err)

	tdirPB := thingdirpb.NewThingDirPB(tdirConfig, hubConfig)
	err = tdirPB.Start()
	assert.NoError(t, err)

	tdirClient := dirclient.NewDirClient(tdirConfig.DirAddress, tdirConfig.DirPort, caCertPath)
	err = tdirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	assert.NoError(t, err)

	_, err = tdirClient.ListTDs(0, 0)
	assert.NoError(t, err)

	tdirClient.Close()
	tdirPB.Stop()
}

func TestUpdateTD(t *testing.T) {
	tdirConfig := &thingdirpb.ThingDirPBConfig{DirAddress: hubConfig.MqttAddress}
	hubConfig, err := hubconfig.LoadCommandlineConfig(homeFolder, thingdirpb.PluginID, &tdirConfig)
	assert.NoError(t, err)

	tdirPB := thingdirpb.NewThingDirPB(tdirConfig, hubConfig)
	err = tdirPB.Start()
	assert.NoError(t, err)

	tdirClient := dirclient.NewDirClient(tdirConfig.DirAddress, tdirConfig.DirPort, caCertPath)
	err = tdirClient.ConnectWithClientCert(pluginCertPath, pluginKeyPath)
	assert.NoError(t, err)

	// Publishing a TD should update the directory
	mbusClient := hubclient.NewMqttHubPluginClient("testUpdateTD", hubConfig)
	mbusClient.Connect()
	require.NotNil(t, mbusClient)
	td1 := td.CreateTD("thing1", vocab.DeviceTypeButton)
	err = mbusClient.PublishTD("thing1", td1)
	assert.NoError(t, err)

	// update takes place in the background so wait a few msec
	time.Sleep(time.Second)
	tds, err := tdirClient.ListTDs(0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(tds))

	tdirClient.Close()
	tdirPB.Stop()
}
