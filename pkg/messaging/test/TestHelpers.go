// Package messaging_test with helper functions
package testhelper

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/gateway/pkg/messaging"
)

const clientID = "Client1"
const protocol = messaging.ConnectionProtocolMQTT

var altServer = ""

// TMessengerConnect helper to test messenger connection
func TMessengerConnect(t *testing.T, gwm messaging.IGatewayMessenger) {
	const timeout = 10
	err := gwm.Connect(clientID, timeout)
	assert.NoError(t, err)
	// reconnect
	err = gwm.Connect(clientID, timeout)
	assert.NoError(t, err)
	gwm.Disconnect()
}
func TMessengerNoConnect(t *testing.T, gwm messaging.IGatewayMessenger) {
	timeout := 5
	require.NotNil(t, gwm)
	err := gwm.Connect(clientID, timeout)
	assert.Error(t, err)
	gwm.Disconnect()
}

func TMessengerPubSub(t *testing.T, gwm messaging.IGatewayMessenger) {
	var rx string
	var msg1 = "Hello world"
	// clientID := "test"
	const timeout = 10
	// certFolder := ""
	err := gwm.Connect(clientID, timeout)
	require.NoError(t, err)

	gwm.Subscribe(messaging.TestChannelID, func(channel string, msg []byte) {
		rx = string(msg)
		logrus.Infof("TMessengerPubSubWithTLS: Received message: %s", msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", messaging.TestChannelID)

	err = gwm.Publish(messaging.TestChannelID, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")

	// allow time to receive
	time.Sleep(1000 * time.Millisecond)
	require.Equalf(t, msg1, rx, "Did not receive the message")

	gwm.Disconnect()
}

// test that only the most recent subscription holds
func TMessengerMultipleSubscriptions(t *testing.T, gwm messaging.IGatewayMessenger) {
	var rx1 string
	var rx2 string
	var msg1 = "Hello world 1"
	var msg2 = "Hello world 2"
	// clientID := "test"
	const timeout = 10

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	err := gwm.Connect(clientID, timeout)
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
	gwm.Subscribe(messaging.TestChannelID, handler1)
	gwm.Subscribe(messaging.TestChannelID, handler2)
	err = gwm.Publish(messaging.TestChannelID, []byte(msg1))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Did not receive the message on handler 1")
	assert.Equalf(t, msg1, rx2, "Did not receive the message on handler 2")

	// after unsubscribe no message should be received by handler 1
	rx1 = ""
	rx2 = ""
	gwm.Unsubscribe(messaging.TestChannelID)
	err = gwm.Publish(messaging.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Received a message on handler 2 after unsubscribe")

	rx1 = ""
	rx2 = ""
	gwm.Unsubscribe(messaging.TestChannelID)
	err = gwm.Publish(messaging.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")

	// when unsubscribing without handler, all handlers should be unsubscribed
	gwm.Subscribe(messaging.TestChannelID, handler1)
	gwm.Subscribe(messaging.TestChannelID, handler2)
	gwm.Unsubscribe(messaging.TestChannelID)
	err = gwm.Publish(messaging.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")

	gwm.Disconnect()
}

func TMessengerBadUnsubscribe(t *testing.T, gwm messaging.IGatewayMessenger) {
	// clientID := "test"
	const timeout = 10

	err := gwm.Connect(clientID, timeout)
	require.NoError(t, err)

	gwm.Unsubscribe(messaging.TestChannelID)
	gwm.Disconnect()

}

func TMessengerPubNoConnect(t *testing.T, gwm messaging.IGatewayMessenger) {
	// clientID := "test"
	const timeout = 10
	var msg1 = "Hello world 1"

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	err := gwm.Publish(messaging.TestChannelID, []byte(msg1))
	require.Error(t, err)

	gwm.Disconnect()

}

func TMessengerSubBeforeConnect(t *testing.T, gwm messaging.IGatewayMessenger) {
	// clientID := "test"
	const timeout = 10
	const msg = "hello 1"
	var rx string
	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	handler1 := func(channel string, msg []byte) {
		logrus.Infof("TestMQTTPubSub: Received message on handler 1: %s", msg)
		rx = string(msg)
	}
	gwm.Subscribe(messaging.TestChannelID, handler1)

	err := gwm.Connect(clientID, timeout)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	err = gwm.Publish(messaging.TestChannelID, []byte(msg))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, msg, rx)

	gwm.Disconnect()

}
