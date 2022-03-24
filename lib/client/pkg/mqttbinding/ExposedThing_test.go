package mqttbinding_test

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
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

	// step 2 create a ExposedThing (why does ConsumedThing uses Consume and ExposedThing Produce?)
	eThing := mqttbinding.Produce(testTD, client)
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

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and ExposedThing
	cThing := mqttbinding.Consume(testTD, client)
	eThing := mqttbinding.Produce(testTD, client)
	eThing.SetPropertyWriteHandler("",
		func(propName string, val mqttbinding.InteractionOutput) error {
			// accept the new value and publish the result
			eThing.EmitPropertyChange(propName, val)
			return nil
		})
	err = eThing.Expose()
	assert.NoError(t, err)
	assert.NotNil(t, eThing)

	// step 3 emit a property change
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

	// step 5 cleanup
	eThing.Destroy()
	client.Close()
}

func TestHandleActionRequest(t *testing.T) {
	logrus.Infof("--- TestInvokeAction ---")
	var receivedAction bool = false

	// step 1 create the MQTT message bus client
	client := mqttclient.NewMqttClient(testPluginID, certs.CaCert, 0)
	err := client.ConnectWithClientCert(mqttCertAddress, certs.PluginCert)
	assert.NoError(t, err)

	// step 2 create a ConsumedThing and ExposedThing
	cThing := mqttbinding.Consume(testTD, client)
	eThing := mqttbinding.Produce(testTD, client)
	eThing.SetActionHandler("",
		func(name string, val mqttbinding.InteractionOutput) error {
			receivedAction = testActionName == name
			return nil
		})
	err = eThing.Expose()
	assert.NoError(t, err)
	assert.NotNil(t, eThing)

	// step 3  an action
	err = cThing.InvokeAction(testActionName, "hi there")
	assert.NoError(t, err)
	time.Sleep(time.Second)

	// step 4 test result. Both exposed and consumed thing must have the new value
	assert.True(t, receivedAction)

	// step 5 cleanup
	eThing.Destroy()
	client.Close()
}
