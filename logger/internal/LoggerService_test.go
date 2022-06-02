package internal_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/wostzone/wost-go/pkg/logging"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/hub/logger/internal"
	"github.com/wostzone/wost-go/pkg/consumedthing"
	"github.com/wostzone/wost-go/pkg/exposedthing"
	"github.com/wostzone/wost-go/pkg/mqttclient"
	"github.com/wostzone/wost-go/pkg/testenv"
	"github.com/wostzone/wost-go/pkg/thing"
	"github.com/wostzone/wost-go/pkg/vocab"
)

const publisherID = "loggerservice"
const testPluginID = "logger-test"

// certificates to test with
var testCerts testenv.TestCerts

// const loremIpsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor " +
// 	"incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco " +
// 	"laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate " +
// 	"velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, " +
// 	"sunt in culpa qui officia deserunt mollit anim id est laborum."

// default configuration for testing
var loggerConfig internal.LoggerServiceConfig

// folder where mosquitto configuration, and logs are generated
var tempFolder string

// TestMain run mosquitto and use the project
// Make sure the certificates exist.
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	testCerts = testenv.CreateCertBundle()
	tempFolder = path.Join(os.TempDir(), "wost-logger-test")
	mosquittoCmd, _ := testenv.StartMosquitto(&testCerts, tempFolder)
	if mosquittoCmd == nil {
		logrus.Fatalf("Unable to setup mosquitto")
	}
	loggerConfig.MqttAddress = testenv.ServerAddress
	loggerConfig.MqttPortCert = testenv.MqttPortCert
	loggerConfig.LogFolder = tempFolder

	result := m.Run()
	_ = mosquittoCmd.Process.Kill()
	if result == 0 {
		os.RemoveAll(tempFolder)
	}

	os.Exit(result)
}

// Test starting and stopping of the logger service
func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")
	config2 := loggerConfig
	config2.ExposeService = true
	svc := internal.NewLoggerService(config2, testCerts.PluginCert, testCerts.CaCert)
	err := svc.Start()
	assert.NoError(t, err)
	svc.Stop()
}

// Test logging of a published TD
func TestLogTD(t *testing.T) {
	logrus.Infof("--- TestLogTD ---")
	deviceID := "device1"
	thingID1 := thing.CreatePublisherID("", publisherID, deviceID, vocab.DeviceTypeSensor)
	//clientID := "TestLogTD"
	eventName1 := "event1"

	config2 := loggerConfig
	config2.ClientID = testPluginID
	config2.ExposeService = true
	svc := internal.NewLoggerService(config2, testCerts.PluginCert, testCerts.CaCert)
	err := svc.Start()
	assert.NoError(t, err)
	// clean start
	logFile := path.Join(config2.LogFolder, thingID1+".log")
	os.Remove(logFile)

	// create a thing to publish with
	etFactory := exposedthing.CreateExposedThingFactory(deviceID, testCerts.DeviceCert, testCerts.CaCert)
	err = etFactory.Connect(config2.MqttAddress, config2.MqttPortCert)
	require.NoError(t, err)

	tdoc := thing.CreateTD(thingID1, "test thing", vocab.DeviceTypeSensor)
	tdoc.UpdateEvent(eventName1, &thing.EventAffordance{})
	eThing, _ := etFactory.Expose(deviceID, tdoc)
	assert.NoError(t, err)

	err = eThing.EmitEvent(eventName1, "test event")
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	etFactory.Destroy(eThing)
	etFactory.Disconnect()

	// verify resulting logfile
	assert.FileExists(t, logFile)

	assert.NoError(t, err)
	svc.Stop()
}

// Test logging of a specific ID
func TestLogSpecificIDs(t *testing.T) {
	logrus.Infof("--- TestLogSpecificIDs ---")
	thingID2 := "urn:zone1:thing2"
	clientID := "TestLogSpecificIDs"
	eventName1 := "event1"
	eventName2 := "event2"

	// load config and start logger
	config2 := loggerConfig
	config2.ClientID = testPluginID
	config2.ThingIDs = []string{thingID2}
	svc := internal.NewLoggerService(config2, testCerts.PluginCert, testCerts.CaCert)
	err := svc.Start()
	assert.NoError(t, err)
	logFile := path.Join(config2.LogFolder, thingID2+".log")
	_ = os.Remove(logFile)

	// create a client to publish events with
	client := mqttclient.NewMqttClient(clientID, testCerts.CaCert, 0)
	hostPort := fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortCert)
	err = client.ConnectWithClientCert(hostPort, testCerts.PluginCert)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// publish the events
	topic1 := strings.ReplaceAll(consumedthing.TopicEmitEvent, "{thingID}", thingID2) + "/" + eventName1
	err = client.PublishObject(topic1, "event1")
	assert.NoError(t, err)

	topic2 := strings.ReplaceAll(consumedthing.TopicEmitEvent, "{thingID}", thingID2) + "/" + eventName2
	err = client.PublishObject(topic2, "event2")
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	client.Disconnect()

	// TODO: verify results
	// verify resulting logfile
	assert.FileExists(t, logFile)

	assert.NoError(t, err)
	svc.Stop()
}

func TestBadLoggingFolder(t *testing.T) {
	logrus.Infof("--- TestBadLoggingFolder ---")
	// load config and start logger
	config2 := loggerConfig
	config2.ClientID = testPluginID
	config2.LogFolder = "/notafolder"
	svc := internal.NewLoggerService(config2, testCerts.PluginCert, testCerts.CaCert)
	err := svc.Start()
	assert.Error(t, err)
	svc.Stop()
}

func TestLogAfterStop(t *testing.T) {
	logrus.Infof("--- TestLogAfterStop ---")
	thingID2 := "urn:zone1:thing2"
	clientID := "TestLogSpecificIDs"
	eventName1 := "event1"

	config2 := loggerConfig
	config2.ClientID = testPluginID
	svc := internal.NewLoggerService(config2, testCerts.PluginCert, testCerts.CaCert)
	err := svc.Start()
	assert.NoError(t, err)
	svc.Stop()

	// create a client to publish events with
	client := mqttclient.NewMqttClient(clientID, testCerts.CaCert, 0)
	hostPort := fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortCert)
	err = client.ConnectWithClientCert(hostPort, testCerts.PluginCert)
	require.NoError(t, err)

	// publish the events
	topic1 := strings.ReplaceAll(consumedthing.TopicEmitEvent, "{thingID}", thingID2) + "/" + eventName1
	err = client.PublishObject(topic1, "event1")
	assert.NoError(t, err)

}
