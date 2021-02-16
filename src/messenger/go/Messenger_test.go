// Package messenger_test with helper functions
package messenger_test

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	messenger "github.com/wostzone/gateway/src/messenger/go"
)

const clientID = "Client1"

// These tests require an server running on hostNamePort
func TMessengerConnect(t *testing.T, client messenger.IGatewayMessenger, hostPort string) {
	const timeout = 10
	err := client.Connect(hostPort, clientID, timeout)
	assert.NoError(t, err)
	// reconnect
	err = client.Connect(hostPort, clientID, timeout)
	assert.NoError(t, err)
	client.Disconnect()
}

func TMessengerNoConnect(t *testing.T, client messenger.IGatewayMessenger) {
	timeout := 5
	require.NotNil(t, client)
	err := client.Connect("invalid.local", clientID, timeout)
	assert.Error(t, err)
	client.Disconnect()
}

func TMessengerPubSubNoTLS(t *testing.T, client messenger.IGatewayMessenger, hostPort string) {
	var rx string
	var msg1 = "Hello world"
	const timeout = 10

	err := client.Connect(hostPort, clientID, timeout)
	require.NoError(t, err)
	client.Subscribe(messenger.TestChannelID, func(channel string, msg []byte) {
		rx = string(msg)
		logrus.Infof("TMessengerPubSubNoTLS: Received message: %s", msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", messenger.TestChannelID)

	err = client.Publish(messenger.TestChannelID, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")

	// allow time to receive
	time.Sleep(1000 * time.Millisecond)
	require.Equalf(t, msg1, rx, "Did not receive the message")

	client.Disconnect()
}

func TMessengerPubSubWithTLS(t *testing.T, client messenger.IGatewayMessenger, hostPort string) {
	var rx string
	var msg1 = "Hello world"
	// clientID := "test"
	const timeout = 10
	// certFolder := ""
	err := client.Connect(hostPort, clientID, timeout)
	require.NoError(t, err)

	client.Subscribe(messenger.TestChannelID, func(channel string, msg []byte) {
		rx = string(msg)
		logrus.Infof("TMessengerPubSubWithTLS: Received message: %s", msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", messenger.TestChannelID)

	err = client.Publish(messenger.TestChannelID, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")

	// allow time to receive
	time.Sleep(1000 * time.Millisecond)
	require.Equalf(t, msg1, rx, "Did not receive the message")

	client.Disconnect()
}

// test that only the most recent subscription holds
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
	client.Subscribe(messenger.TestChannelID, handler1)
	client.Subscribe(messenger.TestChannelID, handler2)
	err = client.Publish(messenger.TestChannelID, []byte(msg1))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Did not receive the message on handler 1")
	assert.Equalf(t, msg1, rx2, "Did not receive the message on handler 2")

	// after unsubscribe no message should be received by handler 1
	rx1 = ""
	rx2 = ""
	client.Unsubscribe(messenger.TestChannelID)
	err = client.Publish(messenger.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Received a message on handler 2 after unsubscribe")

	rx1 = ""
	rx2 = ""
	client.Unsubscribe(messenger.TestChannelID)
	err = client.Publish(messenger.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")

	// when unsubscribing without handler, all handlers should be unsubscribed
	client.Subscribe(messenger.TestChannelID, handler1)
	client.Subscribe(messenger.TestChannelID, handler2)
	client.Unsubscribe(messenger.TestChannelID)
	err = client.Publish(messenger.TestChannelID, []byte(msg2))
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

	client.Unsubscribe(messenger.TestChannelID)
	client.Disconnect()

}

func TMessengerPubNoConnect(t *testing.T, client messenger.IGatewayMessenger) {
	// clientID := "test"
	const timeout = 10
	var msg1 = "Hello world 1"

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	err := client.Publish(messenger.TestChannelID, []byte(msg1))
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
	client.Subscribe(messenger.TestChannelID, handler1)

	err := client.Connect(hostPort, clientID, timeout)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	err = client.Publish(messenger.TestChannelID, []byte(msg))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, msg, rx)

	client.Disconnect()

}
