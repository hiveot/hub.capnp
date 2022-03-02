package mqttclient_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/td"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

const zone = "test"
const testDeviceID = "device1"

//--- THIS USES THE SETUP IN MqttClient_test.go

func TestConnect(t *testing.T) {
	logrus.Infof("--- TestConnect unpw ---")
	consumerClient := mqttclient.NewMqttHubClient(testPluginID, certs.CaCert)
	// the test environment doesn't require a valid password
	err := consumerClient.ConnectWithPassword(mqttUnpwAddress, "username", "badpass")
	assert.NoError(t, err)
	consumerClient.Close()
}

func TestPublishAction(t *testing.T) {
	logrus.Infof("--- TestPublishAction ---")
	thingID := "thing1"
	var rxName string
	var rxParams map[string]interface{}
	actionName := "action1"
	actionInput := map[string]interface{}{"input1": "inputValue"}
	// Use plugin as client with certificate so no hassle with username/pwsswd in testing
	// certificate is for localhost
	consumerClient := mqttclient.NewMqttHubClient(testPluginID, certs.CaCert)
	// consumerClient := mqttclient.NewMqttHubPluginClient("plugin1", hubConfig)
	consumerClient.SetIndent("  ")
	err := consumerClient.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	deviceClient := mqttclient.NewMqttHubClient(testDeviceID, certs.CaCert)
	deviceClient.SubscribeToActions(thingID, func(thingID string, name string, params map[string]interface{}, sender string) {
		logrus.Infof("TestPublishAction: Received action of Thing %s from client %s", thingID, sender)
		rxName = name
		rxParams = params
	})
	err = deviceClient.ConnectWithClientCert(mqttCertAddress, certs.DeviceCert)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond)

	err = consumerClient.PublishAction(thingID, actionName, actionInput)
	assert.NoError(t, err)

	// TODO, check if it was received by the Thing
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, actionName, rxName)
	assert.Equal(t, actionInput, rxParams)

	deviceClient.Close()
	consumerClient.Close()
	// make sure it doest reconnect
	time.Sleep(1 * time.Second)
}

func TestPublishConfig(t *testing.T) {
	logrus.Infof("--- TestPublishConfig ---")
	thingID := "thing1"
	var rx map[string]interface{}
	var rxID string

	config1 := map[string]interface{}{"prop1": "value1"}
	// Use plugin as client with certificate so no hassle with username/pwsswd in testing
	consumerClient := mqttclient.NewMqttHubClient(testPluginID, certs.CaCert)
	err := consumerClient.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	deviceClient := mqttclient.NewMqttHubClient(testDeviceID, certs.CaCert)
	deviceClient.SubscribeToConfig(thingID, func(thingID string, config map[string]interface{}, sender string) {
		logrus.Infof("TestPublishConfig: Received config of Thing %s from client %s", thingID, sender)
		rx = config
		rxID = thingID
	})
	err = deviceClient.ConnectWithClientCert(mqttCertAddress, certs.DeviceCert)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = consumerClient.PublishConfigRequest(thingID, config1)
	assert.NoError(t, err)

	// TODO, check if it was received by the Thing
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, config1["prop1"], rx["prop1"])
	assert.Equal(t, thingID, rxID)
	deviceClient.Close()
	consumerClient.Close()
}

func TestPublishEvent(t *testing.T) {
	logrus.Infof("--- TestPublishEvent ---")
	thingID := "thing1"
	event1 := map[string]interface{}{"eventName": "eventValue"}
	var rx map[string]interface{}

	// Use plugin as client with certificate so no hassle with username/pwsswd in testing
	consumerClient := mqttclient.NewMqttHubClient(testPluginID, certs.CaCert)
	err := consumerClient.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	deviceClient := mqttclient.NewMqttHubClient(testDeviceID, certs.CaCert)
	err = deviceClient.ConnectWithClientCert(mqttCertAddress, certs.DeviceCert)
	assert.NoError(t, err)

	consumerClient.SubscribeToEvents(thingID, func(thingID string, event map[string]interface{}, sender string) {
		logrus.Infof("TestPublishEvent: Received event of Thing %s from client %s", thingID, sender)
		rx = event
	})

	time.Sleep(time.Millisecond)
	err = deviceClient.PublishEvent(thingID, event1)
	assert.NoError(t, err)

	// TODO, check if it was received by a consumer

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, event1["eventName"], rx["eventName"])

	deviceClient.Close()
	consumerClient.Close()
}

func TestPublishPropertyValues(t *testing.T) {
	logrus.Infof("--- TestPublishPropertyValues ---")
	thingID := "thing1"
	propValues := map[string]interface{}{"propname": "value"}
	var rx map[string]interface{}

	// Use plugin as client with certificate so no hassle with username/pwsswd in testing
	consumerClient := mqttclient.NewMqttHubClient(testPluginID, certs.CaCert)
	deviceClient := mqttclient.NewMqttHubClient(testDeviceID, certs.CaCert)

	err := deviceClient.ConnectWithClientCert(mqttCertAddress, certs.DeviceCert)
	assert.NoError(t, err)

	err = consumerClient.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)
	consumerClient.SubscribeToPropertyValues(thingID, func(thingID string, values map[string]interface{}, sender string) {
		logrus.Infof("TestPublishPropertyValues: Received values of Thing %s from client %s", thingID, sender)
		rx = values
	})

	time.Sleep(time.Millisecond)
	err = deviceClient.PublishPropertyValues(thingID, propValues)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, propValues["propname"], rx["propname"])

	deviceClient.Close()
	consumerClient.Close()
}
func TestPublishTD(t *testing.T) {
	logrus.Infof("--- TestPublishTD ---")
	deviceID := "thing1"
	thingID := td.CreateThingID(zone, deviceID, vocab.DeviceTypeSensor)
	td1 := td.CreateTD(thingID, "test TD", vocab.DeviceTypeSensor)
	var rxTd map[string]interface{}

	// Use plugin as client with certificate so no hassle with username/pwsswd in testing
	consumerClient := mqttclient.NewMqttHubClient(testPluginID, certs.CaCert)
	err := consumerClient.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	consumerClient.SubscribeToTD(thingID, func(thingID string, thing map[string]interface{}, sender string) {
		logrus.Infof("TestPublishTD: Received TD of Thing %s from client %s", thingID, sender)
		rxTd = thing
	})

	deviceClient := mqttclient.NewMqttHubClient(testDeviceID, certs.CaCert)
	err = deviceClient.ConnectWithClientCert(mqttCertAddress, certs.DeviceCert)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	err = deviceClient.PublishTD(thingID, td1)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, td1["id"], rxTd["id"])

	// TODO, check if it was received by a consumer using a consumer client
	deviceClient.Close()
	consumerClient.Close()
}

// subscribe to all things
func TestSubscribeAll(t *testing.T) {
	logrus.Infof("--- TestSubscribeAll ---")
	deviceID := "thing1"
	thingID := td.CreateThingID(zone, deviceID, vocab.DeviceTypeSensor)
	td1 := td.CreateTD(thingID, "Test TD", vocab.DeviceTypeSensor)
	txTd, _ := json.MarshalIndent(td1, "  ", "  ")
	var rxTd []byte
	var rxThing string

	deviceClient := mqttclient.NewMqttHubClient(testDeviceID, certs.CaCert)
	err := deviceClient.ConnectWithClientCert(mqttCertAddress, certs.DeviceCert)
	assert.NoError(t, err)

	// Use plugin as client with certificate so no hassle with username/pwsswd in testing
	pluginClient := mqttclient.NewMqttHubClient(testPluginID, certs.CaCert)
	pluginClient.Subscribe("", func(thingID string, msgType string, raw []byte, sender string) {
		logrus.Infof("TestSubscribe: Received msg %s of Thing %s from client %s", msgType, thingID, sender)
		rxTd = raw
		rxThing = thingID
	})
	err = pluginClient.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)

	err = deviceClient.PublishTD(thingID, td1)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, string(txTd), string(rxTd))
	assert.Equal(t, thingID, rxThing)

	// after unsubscribe there should be no more messages
	pluginClient.Unsubscribe("")
	time.Sleep(100 * time.Millisecond)
	err = deviceClient.PublishTD(thingID, td1)
	assert.NoError(t, err)
	rxTd = nil
	time.Sleep(100 * time.Millisecond)
	assert.NotEqual(t, td1, rxTd)

	// TODO, check if it was received by a consumer using a consumer client
	deviceClient.Close()
	pluginClient.Close()
}
