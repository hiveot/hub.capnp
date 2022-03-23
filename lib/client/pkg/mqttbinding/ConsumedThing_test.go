package mqttbinding_test

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

// CA, server and plugin test certificate
var certs testenv.TestCerts

var homeFolder string
var configFolder string

// Use test/mosquitto-test.conf and a client cert port
var mqttCertAddress = fmt.Sprintf("%s:%d", testenv.ServerAddress, testenv.MqttPortCert)

// For running mosquitto in test
const testPluginID = "test-plugin"

const testDeviceID = "device1"
const testDeviceType = vocab.DeviceTypeButton
const testProp1Name = "prop1"
const testProp1Value = "value1"

var testThingID = thing.CreateThingID("", testDeviceID, testDeviceType)
var testTD = createTestTD()

// Create a test TD with
func createTestTD() *thing.ThingTD {
	title := "test Thing"
	thingID := thing.CreateThingID("", testDeviceID, testDeviceType)
	tdDoc := thing.CreateTD(thingID, title, testDeviceType)
	//
	prop := &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type:  vocab.WoTDataTypeBool,
			Title: "Property 1",
		},
	}
	tdDoc.UpdateProperty(testProp1Name, prop)

	return tdDoc
}

// TestMain - launch mosquitto to publish and subscribe
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	configFolder = path.Join(homeFolder, "config")
	certFolder := path.Join(homeFolder, "certs")
	os.Chdir(homeFolder)

	testenv.SetLogging("info", "")
	certs = testenv.CreateCertBundle()
	mosquittoCmd, err := testenv.StartMosquitto(configFolder, certFolder, &certs)
	if err != nil {
		logrus.Fatalf("Unable to start mosquitto: %s", err)
	}

	result := m.Run()
	testenv.StopMosquitto(mosquittoCmd)
	os.Exit(result)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")

	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	ct := mqttbinding.Consume(testTD, client)
	require.NotNil(t, ct)
	assert.Equal(t, testThingID, ct.GetThingDescription().ID)

	ct.Stop()

	assert.NoError(t, err)
	client.Close()
}

func TestReadProperty(t *testing.T) {
	logrus.Infof("--- TestReadProperty ---")
	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing
	cThing := mqttbinding.Consume(testTD, client)

	// step 3 publish the property value (impersonate an ExposedThing)
	topic := strings.ReplaceAll(mqttbinding.TopicThingEvent, "{id}", testThingID) + "/" + testProp1Name
	err = client.PublishObject(topic, testProp1Value)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 4 read the property value. It should match
	val1, err := cThing.ReadProperty(testProp1Name)
	assert.NoError(t, err)
	assert.NotNil(t, val1)

	// step 5 cleanup
	cThing.Stop()
	client.Close()
}

func TestReceiveEvent(t *testing.T) {
	logrus.Infof("--- TestReceiveEvent ---")
	const eventName = "event1"
	const eventValue = "hello world"
	var receivedEvent = false

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and subscribe to event
	cThing := mqttbinding.Consume(testTD, client)
	cThing.SubscribeEvent(eventName, func(ev string, data mqttbinding.InteractionOutput) {
		receivedEvent = eventName == ev
		receivedText := data.ValueAsString()
		assert.Equal(t, eventValue, receivedText)
	})

	// step 3 publish the event (impersonate an ExposedThing)
	topic := strings.ReplaceAll(mqttbinding.TopicThingEvent, "{id}", testThingID) + "/" + eventName
	err = client.PublishObject(topic, eventValue)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 4 check result
	assert.True(t, receivedEvent)

	// step 5 cleanup
	cThing.Stop()
	client.Close()
}
