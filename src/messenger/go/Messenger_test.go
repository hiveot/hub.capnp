package messenger_test

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	messenger "github.com/wostzone/gateway/messenger/go"
)

const clientID = "test"

// These tests require an MQTT TLS server on localhost with TLS support
func TMessengerConnect(t *testing.T, client messenger.IGatewayMessenger, hostPort string) {
	const timeout = 10
	// clientID := "test"
	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	require.NotNil(t, client)

	err := client.Connect(hostPort, clientID, timeout)
	assert.NoError(t, err)

	// reconnect
	err = client.Connect(hostPort, clientID, timeout)
	assert.NoError(t, err)
	client.Disconnect()
}

func TMessengerNoConnect(t *testing.T, client messenger.IGatewayMessenger) {
	// clientID := "test"
	timeout := 5
	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	require.NotNil(t, client)
	err := client.Connect("invalid.local", clientID, timeout)
	assert.Error(t, err)
	// err := gwc.Publish("test1", nil)
	// assert.Error(t, err, "Publish to invalid server should fail")
	client.Disconnect()
}

func TMessengerNoClientID(t *testing.T, client messenger.IGatewayMessenger, hostPort string) {
	// clientID := ""
	timeout := 5
	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	require.NotNil(t, client)
	err := client.Connect(hostPort, clientID, timeout)
	assert.NoError(t, err)
	// err := gwc.Publish("test1", nil)
	// assert.Error(t, err, "Publish to invalid server should fail")
	client.Disconnect()
}
func TMessengerPubSub(t *testing.T, client messenger.IGatewayMessenger, hostPort string) {
	var rx string
	var msg1 = "Hello world"
	// clientID := "test"
	const timeout = 10
	// certFolder := ""

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	err := client.Connect(hostPort, clientID, timeout)
	require.NoError(t, err)
	client.Subscribe(TestChannelID, func(channel string, msg []byte) {
		rx = string(msg)
		logrus.Infof("TestMQTTPubSub: Received message: %s", msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", lib.TestChannelID)

	err = client.Publish(lib.TestChannelID, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")

	// allow time to receive
	time.Sleep(10 * time.Millisecond)
	require.Equalf(t, msg1, rx, "Did not receive the message")

	client.Disconnect()
}

func TMessengerMultipleSubscriptions(t *testing.T, client messenger.IGatewayMessenger, hostPort string) {
	var rx1 string
	var rx2 string
	var msg1 = "Hello world 1"
	var msg2 = "Hello world 2"
	// clientID := "test"
	const timeout = 10

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	err := client.Connect(hostPort, clientID, timeout)
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
	client.Subscribe(lib.TestChannelID, handler1)
	client.Subscribe(lib.TestChannelID, handler2)
	err = client.Publish(lib.TestChannelID, []byte(msg1))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, msg1, rx1, "Did not receive the message on handler 1")
	assert.Equalf(t, msg1, rx2, "Did not receive the message on handler 2")

	// after unsubscribe no message should be received by handler 1
	rx1 = ""
	rx2 = ""
	client.Unsubscribe(lib.TestChannelID)
	err = client.Publish(lib.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, msg2, rx2, "Did not receive the message on handler 2")

	rx1 = ""
	rx2 = ""
	client.Unsubscribe(lib.TestChannelID)
	err = client.Publish(lib.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")

	// when unsubscribing without handler, all handlers should be unsubscribed
	client.Subscribe(lib.TestChannelID, handler1)
	client.Subscribe(lib.TestChannelID, handler2)
	client.Unsubscribe(lib.TestChannelID)
	err = client.Publish(lib.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")

	client.Disconnect()
}

func TMessengerBadUnsubscribe(t *testing.T, client messenger.IGatewayMessenger, hostPort string) {
	// clientID := "test"
	const timeout = 10

	err := client.Connect(hostPort, clientID, timeout)
	require.NoError(t, err)

	client.Unsubscribe(lib.TestChannelID)
	client.Disconnect()

}

func TMessengerPubNoConnect(t *testing.T, client messenger.IGatewayMessenger) {
	// clientID := "test"
	const timeout = 10
	var msg1 = "Hello world 1"

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	err := client.Publish(lib.TestChannelID, []byte(msg1))
	require.Error(t, err)

	client.Disconnect()

}

func TMessengerSubBeforeConnect(t *testing.T, client messenger.IGatewayMessenger, hostPort string) {
	// clientID := "test"
	const timeout = 10
	const msg = "hello 1"
	var rx string
	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	handler1 := func(channel string, msg []byte) {
		logrus.Infof("TestMQTTPubSub: Received message on handler 1: %s", msg)
		rx = string(msg)
	}
	client.Subscribe(lib.TestChannelID, handler1)

	err := client.Connect(hostPort, clientID, timeout)
	require.NoError(t, err)

	err = client.Publish(lib.TestChannelID, []byte(msg))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, msg, rx)

	client.Disconnect()

}
