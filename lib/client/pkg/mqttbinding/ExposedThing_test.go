package mqttbinding_test

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"strings"
	"testing"
	"time"
)

// THIS USES ConsumedThing_test for TestMain and creating a test TD

func TestExpose(t *testing.T) {
	logrus.Infof("--- TestProduce ---")

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ExposedThing (why does ConsumedThing uses Consume and ExposedThing CreateExposedThing?)
	eThing := mqttbinding.CreateExposedThing(testTD, client)
	err = eThing.Expose()
	assert.NoError(t, err)
	assert.NotNil(t, eThing)

	// step 3 cleanup
	eThing.Destroy()
	client.Close()
}

func TestEmitPropertyChange(t *testing.T) {
	logrus.Infof("--- TestEmitPropertyChange ---")
	const newValue = "value 2"
	var propChangeEmitted = false

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and ExposedThing
	cThing := mqttbinding.Consume(testTD, client)
	eThing := mqttbinding.CreateExposedThing(testTD, client)
	eThing.SetPropertyWriteHandler("",
		func(propName string, val mqttbinding.InteractionOutput) error {
			// accept the new value and publish the result
			eThing.EmitPropertyChange(propName, val.Value)

			// and as map
			propMap := map[string]interface{}{propName: val.Value}
			eThing.EmitPropertyChanges(propMap, false)
			propChangeEmitted = true
			return nil
		})
	err = eThing.Expose()
	assert.NoError(t, err)
	assert.NotNil(t, eThing)

	// step 3 request a property change
	err = cThing.WriteProperty(testProp1Name, newValue)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 4 test result. Both exposed and consumed thing must have the new value
	cVal, err := cThing.ReadProperty(testProp1Name)
	assert.NoError(t, err)
	assert.Equal(t, newValue, cVal.ValueAsString())

	eVal, err := eThing.ReadProperty(testProp1Name)
	assert.NoError(t, err)
	assert.Equal(t, newValue, eVal.ValueAsString())

	assert.True(t, propChangeEmitted)

	// step 5 cleanup
	eThing.Destroy()
	client.Close()
}

func TestHandleActionRequest(t *testing.T) {
	logrus.Infof("--- TestHandleActionRequest ---")
	var receivedActionDefaultHandler bool = false
	var receivedActionHandler bool = false

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and ExposedThing with handlers
	testTD.UpdateAction("action2", &thing.ActionAffordance{})
	cThing := mqttbinding.Consume(testTD, client)
	eThing := mqttbinding.CreateExposedThing(testTD, client)
	eThing.SetActionHandler("",
		func(name string, val mqttbinding.InteractionOutput) error {
			receivedActionDefaultHandler = true
			return nil
		})
	eThing.SetActionHandler(testActionName,
		func(name string, val mqttbinding.InteractionOutput) error {
			receivedActionHandler = testActionName == name
			return nil
		})
	err = eThing.Expose()
	assert.NoError(t, err)
	assert.NotNil(t, eThing)
	time.Sleep(time.Second)

	// step 3  an action
	err = cThing.InvokeAction(testActionName, "hi there")
	assert.NoError(t, err)
	time.Sleep(time.Second)
	assert.True(t, receivedActionHandler)
	assert.False(t, receivedActionDefaultHandler)

	// step 4 test result. Both exposed and consumed thing must have the new value
	err = cThing.InvokeAction("action2", nil)
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 4 test result. Both exposed and consumed thing must have the new value
	assert.True(t, receivedActionDefaultHandler)

	// step 5 cleanup
	eThing.Destroy()
	client.Close()
}

func TestHandleActionRequestInvalidParams(t *testing.T) {
	logrus.Infof("--- TestHandleActionRequestInvalidParams ---")
	var receivedAction bool = false

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and ExposedThing with default handler
	eThing := mqttbinding.CreateExposedThing(testTD, client)
	eThing.SetActionHandler("",
		func(name string, val mqttbinding.InteractionOutput) error {
			receivedAction = true
			return nil
		})
	err = eThing.Expose()
	assert.NoError(t, err)
	assert.NotNil(t, eThing)

	// step 3  an action with no name
	topic := strings.ReplaceAll(mqttbinding.TopicInvokeAction, "{thingID}", testThingID)
	err = client.Publish(topic, []byte(testProp1Value))
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 4  an unregistered action
	topic = strings.ReplaceAll(mqttbinding.TopicInvokeAction, "{thingID}", testThingID) + "/badaction"
	err = client.Publish(topic, []byte(testProp1Value))
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 5 test result. no action should have been triggered
	assert.False(t, receivedAction)

	// step 5 cleanup
	eThing.Destroy()
	client.Close()
}

func TestHandleActionRequestNoHandler(t *testing.T) {
	logrus.Infof("--- TestHandleActionRequestNoHandler ---")

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and ExposedThing with default handler
	cThing := mqttbinding.Consume(testTD, client)
	eThing := mqttbinding.CreateExposedThing(testTD, client)
	// no action handler
	err = eThing.Expose()
	assert.NoError(t, err)
	assert.NotNil(t, cThing)
	assert.NotNil(t, eThing)

	// step 3 invoke action with no handler
	err = cThing.InvokeAction(testActionName, "")
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// missing handler does not return an error, just an error in the log

	// step 5 cleanup
	eThing.Destroy()
	client.Close()
}

func TestEmitEventNotConnected(t *testing.T) {
	logrus.Infof("--- TestEmitEventNotConnected ---")

	// step 1 create the MQTT message bus client but dont connect
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	//err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	//assert.NoError(t, err)

	// step 2 create an ExposedThing
	eThing := mqttbinding.CreateExposedThing(testTD, client)
	err := eThing.Expose()
	assert.Error(t, err, "Expect no connection error")

	// step 3 emit event
	err = eThing.EmitEvent(testEventName, "")
	assert.Error(t, err)

	// step 5 cleanup
	eThing.Destroy()
	client.Close()
}

func TestEmitEventNotFound(t *testing.T) {
	logrus.Infof("--- TestEmitEventNotFound ---")

	// step 1 create the MQTT message bus client and connect
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create an ExposedThing
	eThing := mqttbinding.CreateExposedThing(testTD, client)
	err = eThing.Expose()
	assert.NoError(t, err)

	// step 3 emit unknown event
	err = eThing.EmitEvent("unknown-event", "")
	assert.Error(t, err)

	// step 5 cleanup
	eThing.Destroy()
	client.Close()
}
