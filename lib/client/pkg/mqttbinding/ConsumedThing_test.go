package mqttbinding_test

import (
	"encoding/json"
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
	"sync"
	"sync/atomic"
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
const testActionName = "action1"
const testEventName = "event1"
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
	prop1 := &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type:  vocab.WoTDataTypeBool,
			Title: "Property 1",
		},
	}
	prop2 := &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type:  vocab.WoTDataTypeBool,
			Title: "Event property",
		},
	}
	tdDoc.UpdateProperty(testProp1Name, prop1)
	tdDoc.UpdateProperty(testEventName, prop2)

	// add event to TD
	tdDoc.UpdateEvent(testEventName, &thing.EventAffordance{
		Data: thing.DataSchema{},
	})

	// add action to TD
	tdDoc.UpdateAction(testActionName, &thing.ActionAffordance{
		//Input: StringSchema{},
		Safe:       true,
		Idempotent: true,
	})

	return tdDoc
}

// TestMain - launch mosquitto to publish and subscribe
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	configFolder = path.Join(homeFolder, "config")
	certFolder := path.Join(homeFolder, "certs")
	_ = os.Chdir(homeFolder)

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
	var observedProperty int32 = 0

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing
	cThing := mqttbinding.Consume(testTD, client)
	err = cThing.ObserveProperty(testProp1Name, func(name string, data mqttbinding.InteractionOutput) {
		assert.Equal(t, testProp1Name, name)
		atomic.AddInt32(&observedProperty, 1)
	})
	assert.NoError(t, err)

	// step 3 publish the property value (impersonate an ExposedThing)
	//topic := strings.ReplaceAll(mqttbinding.TopicEmitEvent, "{thingID}", testThingID) + "/" + testProp1Name
	topic := mqttbinding.CreateTopic(testThingID, mqttbinding.TopicTypeEvent) + "/" + testProp1Name
	err = client.PublishObject(topic, testProp1Value)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 4 read the property value. It should match
	val1, err := cThing.ReadProperty(testProp1Name)
	assert.NoError(t, err)
	assert.NotNil(t, val1)
	assert.Equal(t, int32(1), atomic.LoadInt32(&observedProperty))

	propNames := []string{testProp1Name}
	propInfo := cThing.ReadMultipleProperties(propNames)
	assert.Equal(t, len(propInfo), 1)

	propInfo = cThing.ReadAllProperties()
	assert.GreaterOrEqual(t, len(propInfo), 1)

	// step 5 cleanup
	cThing.Stop()
	client.Close()
}

func TestReceiveEvent(t *testing.T) {
	logrus.Infof("--- TestReceiveEvent ---")
	const eventName = "event1"
	const eventValue = "hello world"
	var receivedEvent int32 = 0

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and subscribe to event
	cThing := mqttbinding.Consume(testTD, client)
	err = cThing.SubscribeEvent(eventName, func(ev string, data mqttbinding.InteractionOutput) {
		if eventName == ev {
			atomic.AddInt32(&receivedEvent, 1)
		}
		receivedText := data.ValueAsString()
		assert.Equal(t, eventValue, receivedText)
	})
	assert.NoError(t, err)

	// step 3 publish the event (impersonate an ExposedThing)
	topic := strings.ReplaceAll(mqttbinding.TopicEmitEvent, "{thingID}", testThingID) + "/" + eventName
	err = client.PublishObject(topic, eventValue)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 4 check result
	assert.Equal(t, int32(1), atomic.LoadInt32(&receivedEvent))

	// step 5 cleanup
	cThing.Stop()
	client.Close()
}

func TestInvokeAction(t *testing.T) {
	logrus.Infof("--- TestInvokeAction ---")
	const actionValue = "1 2 3 action!"
	var receivedAction int = 0
	var rxMutex = sync.Mutex{}

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and listen for actions on the mqtt bus
	cThing := mqttbinding.Consume(testTD, client)
	actionTopic := strings.ReplaceAll(mqttbinding.TopicInvokeAction, "{thingID}", testThingID) + "/#"
	client.Subscribe(actionTopic, func(address string, message []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		receivedAction++
		var rxData2 string
		err := json.Unmarshal(message, &rxData2)
		assert.NoError(t, err)
		assert.Equal(t, actionValue, rxData2)
	})

	// step 3 publish the action
	err = cThing.InvokeAction(testActionName, actionValue)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 4 check result
	rxMutex.Lock()
	assert.Equal(t, 1, receivedAction)
	defer rxMutex.Unlock()

	// step 5 cleanup
	cThing.Stop()
	client.Unsubscribe(actionTopic)
	client.Close()
}

func TestInvokeActionBadName(t *testing.T) {
	logrus.Infof("--- TestInvokeActionBadName ---")

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and listen for actions on the mqtt bus
	cThing := mqttbinding.Consume(testTD, client)

	// step 3 publish the action with unknown name
	err = cThing.InvokeAction("unknown-action", "")
	assert.Error(t, err)

	// step 4 cleanup
	cThing.Stop()
	client.Close()
}

func TestWriteProperty(t *testing.T) {
	const testNewPropValue1 = "new value 1"
	const testNewPropValue2 = "new value 2"
	logrus.Infof("--- TestWriteProperty ---")

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing
	cThing := mqttbinding.Consume(testTD, client)

	// step 3 submit the write request
	err = cThing.WriteProperty(testProp1Name, testNewPropValue1)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	newProps := make(map[string]interface{})
	newProps[testProp1Name] = testNewPropValue2
	err = cThing.WriteMultipleProperties(newProps)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 5 cleanup
	cThing.Stop()
	client.Close()
}
