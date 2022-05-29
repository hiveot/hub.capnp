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
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/exposedthing"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/mqttclient"
	"github.com/wostzone/wost-go/pkg/testenv"
	"github.com/wostzone/wost-go/pkg/thing"
	"github.com/wostzone/wost-go/pkg/vocab"

	"github.com/sirupsen/logrus"
)

// var certFolder string
var testCerts testenv.TestCerts

// var serverAddress string

var tempFolder string
var hubConfig config.HubConfig

// TestMain runs a directory server for use by the test cases in this package
// This uses the directory client in testing
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of DirectoryServer ------")
	// serverAddress = hubnet.GetOutboundIP("").String()

	cwd, _ := os.Getwd()
	configFolder := path.Join(cwd, "../../test", "config")
	//_ = os.Chdir(appFolder)

	logging.SetLogging("info", "")
	testCerts = testenv.CreateCertBundle()
	tempFolder = path.Join(os.TempDir(), "wost-thingdir-test")
	mosquittoCmd, err := testenv.StartMosquitto(&testCerts, tempFolder)
	if err != nil {
		logrus.Fatalf("Unable to start mosquitto: %s", err)
	}
	time.Sleep(time.Millisecond * 100)

	configFile := path.Join(configFolder, "hub.yaml")
	hubConfig = *config.CreateHubConfig(tempFolder)
	// override config folder to use the test config
	hubConfig.ConfigFolder = configFolder
	// FIXME: should not need to use the clientID "plugin" in order to use the plugin client certificate
	err = hubConfig.Load(configFile, "plugin")
	if err != nil {
		logrus.Fatalf("Unable to load hub config: %s", err)
	}

	res := m.Run()

	//_ = mosquittoCmd.Process.Kill()
	testenv.StopMosquitto(mosquittoCmd, tempFolder)

	os.Exit(res)
}

func TestStartStopThingDirectoryService(t *testing.T) {
	tdirConfig := &thingdirpb.ThingDirPBConfig{}
	configFile := path.Join(hubConfig.ConfigFolder, thingdirpb.PluginID+".yaml")
	err := config.LoadYamlConfig(configFile, &tdirConfig, nil)

	//hubConfig, err := config.LoadCommandlineConfig(homeFolder, thingdirpb.PluginID, &tdirConfig)
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
	mqttClient := mqttclient.NewMqttClient("testUpdateTD", hubConfig.CaCert, 0)
	_ = mqttClient.ConnectWithClientCert(mqttHostPort, hubConfig.PluginCert)
	require.NotNil(t, mqttClient)
	td1 := thing.CreateTD("thing1", "test thing", vocab.DeviceTypeButton)
	err = mqttClient.PublishObject("thing1", td1.AsMap())
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

func TestUpdatePropValues(t *testing.T) {
	const thing1ID = "thing1"
	const prop1Name = "prop1"
	const prop1Value = "value1"
	const event1Name = "event1"
	const event1Value = "eventValue1"

	tdirConfig := &thingdirpb.ThingDirPBConfig{DirAddress: hubConfig.Address}
	configFile := path.Join(hubConfig.ConfigFolder, thingdirpb.PluginID+".yaml")
	err := config.LoadYamlConfig(configFile, &tdirConfig, nil)

	tdirPB := thingdirpb.NewThingDirPB(tdirConfig, &hubConfig)
	err = tdirPB.Start()
	assert.NoError(t, err)

	dirHostPort := fmt.Sprintf("%s:%d", tdirConfig.DirAddress, tdirConfig.DirPort)
	tdirClient := dirclient.NewDirClient(dirHostPort, hubConfig.CaCert)
	err = tdirClient.ConnectWithClientCert(hubConfig.PluginCert)
	assert.NoError(t, err)

	// Publishing a TD should update the directory
	mqttHostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortCert)
	mqttClient := mqttclient.NewMqttClient("TestUpdatePropValues", hubConfig.CaCert, 0)
	mqttClient.ConnectWithClientCert(mqttHostPort, hubConfig.PluginCert)
	require.NotNil(t, mqttClient)

	// use the exposed thing for both TD and values
	td1 := thing.CreateTD(thing1ID, "test thing", vocab.DeviceTypeButton)
	td1.UpdateProperty(prop1Name, &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{Type: vocab.WoTDataTypeString},
	})
	td1.UpdateEvent(event1Name, &thing.EventAffordance{
		Data: thing.DataSchema{Type: vocab.WoTDataTypeString},
	})
	factory := exposedthing.CreateExposedThingFactory("thingdir-test", testCerts.DeviceCert, testCerts.CaCert)
	factory.Connect(hubConfig.Address, hubConfig.MqttPortCert)
	eThing := factory.Expose("device1", td1)
	assert.NotNil(t, eThing)

	// finally update a property and lets include an event
	err = eThing.EmitPropertyChange(prop1Name, prop1Value)
	assert.NoError(t, err)
	err = eThing.EmitEvent(event1Name, event1Value)
	assert.NoError(t, err)

	// update takes place in the background so wait a few msec
	time.Sleep(time.Second)

	// match results
	thingValues, err := tdirClient.GetThingValues(thing1ID)
	assert.NoError(t, err)
	require.NotNil(t, thingValues)
	assert.Equal(t, prop1Value, thingValues[prop1Name].Value)
	assert.Equal(t, event1Value, thingValues[event1Name].Value)

	// cleanup
	factory.Disconnect()
	tdirClient.Close()
	tdirPB.Stop()
}
