package plugin

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mqqtServer = "localhost:8883"
const mqttCertFolder = "/etc/mosquitto/certs"

// These tests require an MQTT TLS server on localhost with TLS support
func TestMqttConnect(t *testing.T) {
	const timeout = 10
	clientID := "test"
	mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	require.NotNil(t, mqttMessenger)

	err := mqttMessenger.Connect(mqqtServer, timeout)
	assert.NoError(t, err)

	// reconnect
	err = mqttMessenger.Connect(mqqtServer, timeout)
	assert.NoError(t, err)
	mqttMessenger.Disconnect()
}

func TestMqttNoConnect(t *testing.T) {
	clientID := "test"
	timeout := 5
	mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	require.NotNil(t, mqttMessenger)
	err := mqttMessenger.Connect("invalid.local", timeout)
	assert.Error(t, err)
	// err := gwc.Publish("test1", nil)
	// assert.Error(t, err, "Publish to invalid server should fail")
	mqttMessenger.Disconnect()
}

func TestMqttNoClientID(t *testing.T) {
	clientID := ""
	timeout := 5
	mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	require.NotNil(t, mqttMessenger)
	err := mqttMessenger.Connect(mqqtServer, timeout)
	assert.NoError(t, err)
	// err := gwc.Publish("test1", nil)
	// assert.Error(t, err, "Publish to invalid server should fail")
	mqttMessenger.Disconnect()
}
func TestMQTTPubSub(t *testing.T) {
	var rx string
	var msg1 = "Hello world"
	clientID := "test"
	const timeout = 10
	// certFolder := ""

	mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	err := mqttMessenger.Connect(mqqtServer, timeout)
	require.NoError(t, err)
	err = mqttMessenger.Subscribe(TestChannelID, func(channel string, msg []byte) {
		rx = string(msg)
		logrus.Infof("TestMQTTPubSub: Received message: %s", msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", TestChannelID)

	err = mqttMessenger.Publish(TestChannelID, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")

	// allow time to receive
	time.Sleep(10 * time.Millisecond)
	require.Equalf(t, msg1, rx, "Did not receive the message")

	mqttMessenger.Disconnect()
}

func TestMQTTMultipleSubscriptions(t *testing.T) {
	var rx1 string
	var rx2 string
	var msg1 = "Hello world 1"
	var msg2 = "Hello world 2"
	clientID := "test"
	const timeout = 10

	mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	err := mqttMessenger.Connect(mqqtServer, timeout)
	require.NoError(t, err)
	handler1 := func(channel string, msg []byte) {
		rx1 = string(msg)
		logrus.Infof("TestMQTTPubSub: Received message on handler 1: %s", msg)
	}
	handler2 := func(channel string, msg []byte) {
		rx2 = string(msg)
		logrus.Infof("TestMQTTPubSub: Received message on handler 2: %s", msg)
	}
	_ = handler2
	err = mqttMessenger.Subscribe(TestChannelID, handler1)
	err = mqttMessenger.Subscribe(TestChannelID, handler2)
	err = mqttMessenger.Publish(TestChannelID, []byte(msg1))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, msg1, rx1, "Did not receive the message on handler 1")
	assert.Equalf(t, msg1, rx2, "Did not receive the message on handler 2")

	// after unsubscribe no message should be received by handler 1
	rx1 = ""
	rx2 = ""
	mqttMessenger.Unsubscribe(TestChannelID, handler1)
	err = mqttMessenger.Publish(TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, msg2, rx2, "Did not receive the message on handler 2")

	rx1 = ""
	rx2 = ""
	mqttMessenger.Unsubscribe(TestChannelID, handler2)
	err = mqttMessenger.Publish(TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")

	// when unsubscribing without handler, all handlers should be unsubscribed
	err = mqttMessenger.Subscribe(TestChannelID, handler1)
	err = mqttMessenger.Subscribe(TestChannelID, handler2)
	mqttMessenger.Unsubscribe(TestChannelID, nil)
	err = mqttMessenger.Publish(TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")

	mqttMessenger.Disconnect()
}

func TestMQTTBadUnsubscribe(t *testing.T) {
	clientID := "test"
	const timeout = 10
	mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	handler1 := func(channel string, msg []byte) {
		logrus.Infof("TestMQTTPubSub: Received message on handler 1: %s", msg)
	}

	err := mqttMessenger.Connect(mqqtServer, timeout)
	require.NoError(t, err)

	mqttMessenger.Unsubscribe(TestChannelID, handler1)
	mqttMessenger.Disconnect()

}

func TestMQTTPubNoConnect(t *testing.T) {
	clientID := "test"
	const timeout = 10
	var msg1 = "Hello world 1"

	mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	err := mqttMessenger.Publish(TestChannelID, []byte(msg1))
	require.Error(t, err)

	mqttMessenger.Disconnect()

}

func TestMQTTSubBeforeConnect(t *testing.T) {
	clientID := "test"
	const timeout = 10
	const msg = "hello 1"
	var rx string
	mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	handler1 := func(channel string, msg []byte) {
		logrus.Infof("TestMQTTPubSub: Received message on handler 1: %s", msg)
		rx = string(msg)
	}
	err := mqttMessenger.Subscribe(TestChannelID, handler1)
	require.NoError(t, err)

	err = mqttMessenger.Connect(mqqtServer, timeout)
	require.NoError(t, err)

	err = mqttMessenger.Publish(TestChannelID, []byte(msg))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, msg, rx)

	mqttMessenger.Disconnect()

}
