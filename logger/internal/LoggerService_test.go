package internal_test

import (
	"fmt"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"github.com/wostzone/hub/logger/internal"
)

var homeFolder string

const zone = "test"
const publisherID = "loggerservice"
const testPluginID = "logger-test"

var testCerts testenv.TestCerts

// const loremIpsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor " +
// 	"incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco " +
// 	"laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate " +
// 	"velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, " +
// 	"sunt in culpa qui officia deserunt mollit anim id est laborum."

var hubConfig *config.HubConfig

var mosquittoCmd *exec.Cmd

// TestMain run mosquitto and use the project test folder as the home folder.
// Make sure the certificates exist.
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")

	hubConfig = config.CreateDefaultHubConfig(homeFolder)
	config.LoadHubConfig("", testPluginID, hubConfig)

	testCerts = testenv.CreateCertBundle()
	testenv.SaveCerts(&testCerts, hubConfig.CertsFolder)

	mosquittoCmd, _ = testenv.StartMosquitto(hubConfig.ConfigFolder, hubConfig.CertsFolder, &testCerts)
	if mosquittoCmd == nil {
		logrus.Fatalf("Unable to setup mosquitto")
	}

	result := m.Run()
	mosquittoCmd.Process.Kill()

	os.Exit(result)
}

// Test starting and stopping of the logger service
func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")

	svc := internal.NewLoggerService()
	svc.Config.PublishTD = true
	hubConfig, err := config.LoadAllConfig(nil, homeFolder, testPluginID, &svc.Config)
	assert.NoError(t, err)
	err = svc.Start(hubConfig)
	assert.NoError(t, err)
	svc.Stop()
}

// Test logging of a published TD
func TestLogTD(t *testing.T) {
	logrus.Infof("--- TestLogTD ---")
	deviceID := "device1"
	thingID1 := thing.CreatePublisherID(zone, publisherID, deviceID, vocab.DeviceTypeSensor)
	clientID := "TestLogTD"
	eventName1 := "event1"

	svc := internal.NewLoggerService()
	hubConfig, err := config.LoadAllConfig(nil, homeFolder, testPluginID, &svc.Config)
	assert.NoError(t, err)
	err = svc.Start(hubConfig)
	assert.NoError(t, err)
	// clean start
	logFile := path.Join(svc.Config.LogsFolder, thingID1+".log")
	os.Remove(logFile)

	client := mqttclient.NewMqttClient(clientID, testCerts.CaCert, 0)
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Address, testenv.MqttPortCert)
	err = client.ConnectWithClientCert(hostPort, hubConfig.PluginCert)
	require.Nil(t, err)
	time.Sleep(100 * time.Millisecond)

	// create a thing to publish with
	tdoc := thing.CreateTD(thingID1, "test thing", vocab.DeviceTypeSensor)
	tdoc.UpdateEvent(eventName1, &thing.EventAffordance{})
	eThing := mqttbinding.CreateExposedThing(tdoc, client)
	err = eThing.Expose()
	assert.NoError(t, err)

	err = eThing.EmitEvent(eventName1, "test event")
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	eThing.Stop()
	client.Close()

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
	svc := internal.NewLoggerService()
	hubConfig, err := config.LoadAllConfig(nil, homeFolder, testPluginID, &svc.Config)
	assert.NoError(t, err)
	svc.Config.ThingIDs = []string{thingID2}
	err = svc.Start(hubConfig)
	assert.NoError(t, err)
	logFile := path.Join(svc.Config.LogsFolder, thingID2+".log")
	os.Remove(logFile)

	// create a client to publish events with
	client := mqttclient.NewMqttClient(clientID, testCerts.CaCert, 0)
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Address, testenv.MqttPortCert)
	err = client.ConnectWithClientCert(hostPort, hubConfig.PluginCert)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// publish the events
	topic1 := strings.ReplaceAll(mqttbinding.TopicThingEvent, "{thingID}", thingID2) + "/" + eventName1
	err = client.PublishObject(topic1, "event1")
	assert.NoError(t, err)

	topic2 := strings.ReplaceAll(mqttbinding.TopicThingEvent, "{thingID}", thingID2) + "/" + eventName2
	err = client.PublishObject(topic2, "event2")
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	client.Close()

	// TODO: verify results
	// verify resulting logfile
	assert.FileExists(t, logFile)

	assert.NoError(t, err)
	svc.Stop()
}

func TestAltLoggingFolder(t *testing.T) {
	logrus.Infof("--- TestAltLoggingFolder ---")

	svc := internal.NewLoggerService()
	hubConfig, err := config.LoadAllConfig(nil, homeFolder, testPluginID, &svc.Config)
	assert.NoError(t, err)
	svc.Config.LogsFolder = "/tmp"
	err = svc.Start(hubConfig)
	assert.NoError(t, err)
	svc.Stop()
}

func TestBadLoggingFolder(t *testing.T) {
	logrus.Infof("--- TestBadLoggingFolder ---")
	svc := internal.NewLoggerService()
	hubConfig, err := config.LoadAllConfig(nil, homeFolder, testPluginID, &svc.Config)
	assert.NoError(t, err)
	svc.Config.LogsFolder = "/notafolder"
	err = svc.Start(hubConfig)
	assert.Error(t, err)
	svc.Stop()
}

func TestLogAfterStop(t *testing.T) {
	logrus.Infof("--- TestLogAfterStop ---")
	svc := internal.NewLoggerService()
	hubConfig, err := config.LoadAllConfig(nil, homeFolder, testPluginID, &svc.Config)
	assert.NoError(t, err)
	err = svc.Start(hubConfig)
	assert.NoError(t, err)

	svc.Stop()
}
