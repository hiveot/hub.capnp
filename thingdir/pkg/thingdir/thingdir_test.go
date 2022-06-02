package thingdir_test

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/hub/authz/pkg/aclstore"
	"github.com/wostzone/hub/thingdir/pkg/dirclient"
	"github.com/wostzone/hub/thingdir/pkg/thingdir"
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
var tempFolder string
var aclFilePath string
var thingDirConfig thingdir.ThingDirConfig

const testDirectoryPort = 9990

// TestMain setup of a test environment for running the directory server
// This uses the directory client in testing
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	logrus.Infof("------ TestMain of DirectoryServer ------")

	// tempFolder holds directory and acl files
	tempFolder = path.Join(os.TempDir(), "wost-thingdir-test")
	os.Mkdir(tempFolder, 0700)

	// clean acls
	aclFilePath = path.Join(tempFolder, aclstore.DefaultAclFile)
	fp, _ := os.Create(aclFilePath)
	_ = fp.Close()

	testCerts = testenv.CreateCertBundle()
	thingDirConfig = thingdir.ThingDirConfig{
		InstanceID:      "thingdir-test",
		DirAddress:      testenv.ServerAddress,
		DirPort:         testDirectoryPort,
		DirAclFile:      aclFilePath,
		DirStoreFolder:  tempFolder,
		EnableDiscovery: false,
		MsgbusAddress:   testenv.ServerAddress,
		MsgbusPortCert:  testenv.MqttPortCert,
	}

	mosquittoCmd, err := testenv.StartMosquitto(&testCerts, tempFolder)
	if err != nil {
		logrus.Fatalf("Unable to start mosquitto: %s", err)
	}
	time.Sleep(time.Millisecond * 100)

	res := m.Run()

	//_ = mosquittoCmd.Process.Kill()
	testenv.StopMosquitto(mosquittoCmd, tempFolder)
	if res == 0 {
		os.RemoveAll(tempFolder)
	}
	os.Exit(res)
}

func TestStartStopThingDirectoryService(t *testing.T) {
	config2 := thingDirConfig

	svc := thingdir.NewThingDir(&config2, testCerts.CaCert, testCerts.ServerCert, testCerts.PluginCert)
	err := svc.Start()
	assert.NoError(t, err)

	dirHostPort := fmt.Sprintf("%s:%d", config2.DirAddress, config2.DirPort)
	tdirClient := dirclient.NewDirClient(dirHostPort, testCerts.CaCert)
	err = tdirClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	_, err = tdirClient.ListTDs(0, 0)
	assert.NoError(t, err)
	logrus.Infof("TestUpdateTD: Closing ")

	tdirClient.Close()
	svc.Stop()
}

func TestStartThingDirBadAddress(t *testing.T) {
	config2 := thingDirConfig
	config2.DirAddress = "wrongaddress"

	svc := thingdir.NewThingDir(&config2, testCerts.CaCert, testCerts.ServerCert, testCerts.PluginCert)
	err := svc.Start()
	assert.Error(t, err)
}

func TestUpdateTD(t *testing.T) {
	const device1ID = "device1"
	var thing1ID = thing.CreateThingID("", device1ID, vocab.DeviceTypeButton)
	config2 := thingDirConfig

	svc := thingdir.NewThingDir(&config2, testCerts.CaCert, testCerts.ServerCert, testCerts.PluginCert)
	err := svc.Start()
	assert.NoError(t, err)

	dirHostPort := fmt.Sprintf("%s:%d", config2.DirAddress, config2.DirPort)
	tdirClient := dirclient.NewDirClient(dirHostPort, testCerts.CaCert)
	err = tdirClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	// Publishing a TD should update the directory
	eFactory := exposedthing.CreateExposedThingFactory("thingdir-test", testCerts.DeviceCert, testCerts.CaCert)
	eFactory.Connect(config2.MsgbusAddress, config2.MsgbusPortCert)
	td1 := thing.CreateTD(thing1ID, "test thing", vocab.DeviceTypeButton)
	eThing, _ := eFactory.Expose(device1ID, td1)
	assert.NotNil(t, eThing)

	// update takes place in the background so wait a few msec
	time.Sleep(time.Second)
	tds, err := tdirClient.ListTDs(0, 0)
	assert.NoError(t, err)
	assert.Greater(t, len(tds), 0, "missing TDs in store")

	logrus.Infof("TestUpdateTD: Closing ")
	tdirClient.Close()
	svc.Stop()
}

func TestUpdatePropValues(t *testing.T) {
	const device1ID = "device1"
	const prop1Name = "prop1"
	const prop1Value = "value1"
	const event1Name = "event1"
	const event1Value = "eventValue1"
	config2 := thingDirConfig
	var thing1ID = thing.CreateThingID("", device1ID, vocab.DeviceTypeButton)

	svc := thingdir.NewThingDir(&config2, testCerts.CaCert, testCerts.ServerCert, testCerts.PluginCert)
	err := svc.Start()
	assert.NoError(t, err)

	dirHostPort := fmt.Sprintf("%s:%d", config2.DirAddress, config2.DirPort)
	tdirClient := dirclient.NewDirClient(dirHostPort, testCerts.CaCert)
	err = tdirClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	// Publishing a TD should update the directory
	mqttHostPort := fmt.Sprintf("%s:%d", config2.MsgbusAddress, config2.MsgbusPortCert)
	mqttClient := mqttclient.NewMqttClient("TestUpdatePropValues", testCerts.CaCert, 0)
	err = mqttClient.ConnectWithClientCert(mqttHostPort, testCerts.PluginCert)
	assert.NoError(t, err)

	// use the exposed thing for updating both TD and values
	td1 := thing.CreateTD(thing1ID, "test thing", vocab.DeviceTypeButton)
	td1.UpdateProperty(prop1Name, &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{Type: vocab.WoTDataTypeString},
	})
	td1.UpdateEvent(event1Name, &thing.EventAffordance{
		Data: thing.DataSchema{Type: vocab.WoTDataTypeString},
	})
	factory := exposedthing.CreateExposedThingFactory("thingdir-test", testCerts.DeviceCert, testCerts.CaCert)
	err = factory.Connect(config2.MsgbusAddress, config2.MsgbusPortCert)
	assert.NoError(t, err)
	eThing, _ := factory.Expose("device1", td1)
	assert.NotNil(t, eThing)

	// finally update a property and lets include an event
	err = eThing.EmitPropertyChange(prop1Name, prop1Value, false)
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
	svc.Stop()
}
