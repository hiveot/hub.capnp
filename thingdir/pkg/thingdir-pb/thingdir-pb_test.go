package thingdirpb_test

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/thingdir/pkg/dirclient"
	thingdirpb "github.com/wostzone/hub/thingdir/pkg/thingdir-pb"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/td"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

// var certFolder string
var testCerts testenv.TestCerts

// var serverAddress string

var appFolder string
var hubConfig config.HubConfig

// TestMain runs a directory server for use by the test cases in this package
// This uses the directory client in testing
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of DirectoryServer ------")
	// serverAddress = hubnet.GetOutboundIP("").String()

	cwd, _ := os.Getwd()
	appFolder = path.Join(cwd, "../../test")
	configFolder := path.Join(appFolder, "config")
	certFolder := path.Join(appFolder, "certs")
	os.Chdir(appFolder)

	testenv.SetLogging("info", "")
	testCerts = testenv.CreateCertBundle()
	mosquittoCmd, err := testenv.StartMosquitto(configFolder, certFolder, &testCerts)
	if err != nil {
		logrus.Fatalf("Unable to start mosquitto: %s", err)
	}

	hubConfig = *config.CreateDefaultHubConfig(appFolder)
	configFile := path.Join(configFolder, "hub.yaml")
	_ = config.LoadHubConfig(configFile, thingdirpb.PluginID, &hubConfig)

	res := m.Run()

	mosquittoCmd.Process.Kill()

	os.Exit(res)
}

func TestStartStopThingDirectoryService(t *testing.T) {
	tdirConfig := &thingdirpb.ThingDirPBConfig{}
	configFile := path.Join(hubConfig.ConfigFolder, thingdirpb.PluginID+".yaml")
	err := config.LoadYamlConfig(configFile, &tdirConfig, nil)

	// hubConfig, err := config.LoadCommandlineConfig(homeFolder, thingdirpb.PluginID, &tdirConfig)
	assert.NoError(t, err)
	tdirPB := thingdirpb.NewThingDirPB(tdirConfig, &hubConfig)
	err = tdirPB.Start()
	assert.NoError(t, err)

	dirHostPort := fmt.Sprintf("%s:%d", testenv.ServerAddress, tdirConfig.DirPort)
	tdirClient := dirclient.NewDirClient(dirHostPort, hubConfig.CaCert)
	err = tdirClient.ConnectWithClientCert(hubConfig.PluginCert)
	assert.NoError(t, err)

	_, err = tdirClient.ListTDs(0, 0)
	assert.NoError(t, err)
	logrus.Infof("TestUpdateTD: Closing ")

	tdirClient.Close()
	tdirPB.Stop()
}

func TestStartThingDirBadAddress(t *testing.T) {
	tdirConfig := &thingdirpb.ThingDirPBConfig{}
	hc := hubConfig // copy
	hc.Address = "wrongaddress"

	tdirPB := thingdirpb.NewThingDirPB(tdirConfig, &hc)
	err := tdirPB.Start()
	assert.Error(t, err)
}

func TestUpdateTD(t *testing.T) {
	tdirConfig := &thingdirpb.ThingDirPBConfig{DirAddress: hubConfig.Address}
	configFile := path.Join(hubConfig.ConfigFolder, thingdirpb.PluginID+".yaml")
	err := config.LoadYamlConfig(configFile, &tdirConfig, nil)

	// hubConfig, err := config.LoadHubConfig("", homeFolder, thingdirpb.PluginID)
	assert.NoError(t, err)
	// err = config.LoadPluginConfig(hubConfig.ConfigFolder, thingdirpb.PluginID, &tdirConfig, nil)

	// hubConfig, err := config.LoadCommandlineConfig(homeFolder, thingdirpb.PluginID, &tdirConfig)
	assert.NoError(t, err)

	tdirPB := thingdirpb.NewThingDirPB(tdirConfig, &hubConfig)
	err = tdirPB.Start()
	assert.NoError(t, err)

	dirHostPort := fmt.Sprintf("%s:%d", tdirConfig.DirAddress, tdirConfig.DirPort)
	tdirClient := dirclient.NewDirClient(dirHostPort, hubConfig.CaCert)
	err = tdirClient.ConnectWithClientCert(hubConfig.PluginCert)
	assert.NoError(t, err)

	// Publishing a TD should update the directory
	mqttHostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortCert)
	mqttClient := mqttclient.NewMqttHubClient("testUpdateTD", hubConfig.CaCert)
	mqttClient.ConnectWithClientCert(mqttHostPort, hubConfig.PluginCert)
	require.NotNil(t, mqttClient)
	td1 := td.CreateTD("thing1", "test thing", vocab.DeviceTypeButton)
	err = mqttClient.PublishTD("thing1", td1)
	assert.NoError(t, err)

	// update takes place in the background so wait a few msec
	time.Sleep(time.Second)
	tds, err := tdirClient.ListTDs(0, 0)
	assert.NoError(t, err)
	assert.Greater(t, len(tds), 0, "missing TDs in store")

	logrus.Infof("TestUpdateTD: Closing ")
	tdirClient.Close()
	tdirPB.Stop()
}
