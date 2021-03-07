// Package testhelper_test with helper functions
package testhelper_test

import (
	"sync"
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

// TMessengerNoConnect helper to test  messenger connection
func TMessengerNoConnect(t *testing.T, gwm messaging.IGatewayMessenger) {
	timeout := 5
	require.NotNil(t, gwm)
	err := gwm.Connect(clientID, timeout)
	assert.Error(t, err)
	gwm.Disconnect()
}

// TMessengerPubSub is a test helper for publish/subscribe a message
func TMessengerPubSub(t *testing.T, gwm messaging.IGatewayMessenger) {
	var rx string
	rxMutex := sync.Mutex{}
	var msg1 = "Hello world"
	// clientID := "test"
	const timeout = 10
	// certFolder := ""
	err := gwm.Connect(clientID, timeout)
	require.NoError(t, err)

	gwm.Subscribe(messaging.TestChannelID, func(channel string, msg []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx = string(msg)
		logrus.Infof("Received message: %s", msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", messaging.TestChannelID)

	err = gwm.Publish(messaging.TestChannelID, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")

	// allow time to receive
	time.Sleep(1000 * time.Millisecond)
	rxMutex.Lock()
	defer rxMutex.Unlock()
	require.Equalf(t, msg1, rx, "Did not receive the message")

	gwm.Disconnect()
}

// TMessengerMultipleSubscriptions test that only the most recent subscription holds
func TMessengerMultipleSubscriptions(t *testing.T, gwm messaging.IGatewayMessenger) {
	var rx1 string
	var rx2 string
	rxMutex := sync.Mutex{}
	var msg1 = "Hello world 1"
	var msg2 = "Hello world 2"
	// clientID := "test"
	const timeout = 10

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)
	err := gwm.Connect(clientID, timeout)
	require.NoError(t, err)
	handler1 := func(channel string, msg []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx1 = string(msg)
		logrus.Infof("Received message on handler 1: %s", msg)
	}
	handler2 := func(channel string, msg []byte) {
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx2 = string(msg)
		logrus.Infof("Received message on handler 2: %s", msg)
	}
	_ = handler2
	gwm.Subscribe(messaging.TestChannelID, handler1)
	gwm.Subscribe(messaging.TestChannelID, handler2)
	err = gwm.Publish(messaging.TestChannelID, []byte(msg1))
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, "", rx1, "Did not receive the message on handler 1")
	assert.Equalf(t, msg1, rx2, "Did not receive the message on handler 2")
	// after unsubscribe no message should be received by handler 1
	rx1 = ""
	rx2 = ""
	rxMutex.Unlock()
	gwm.Unsubscribe(messaging.TestChannelID)
	err = gwm.Publish(messaging.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Received a message on handler 2 after unsubscribe")
	rx1 = ""
	rx2 = ""
	rxMutex.Unlock()

	gwm.Unsubscribe(messaging.TestChannelID)
	err = gwm.Publish(messaging.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")
	rxMutex.Unlock()

	// when unsubscribing without handler, all handlers should be unsubscribed
	gwm.Subscribe(messaging.TestChannelID, handler1)
	gwm.Subscribe(messaging.TestChannelID, handler2)
	gwm.Unsubscribe(messaging.TestChannelID)
	err = gwm.Publish(messaging.TestChannelID, []byte(msg2))
	time.Sleep(100 * time.Millisecond)

	rxMutex.Lock()
	assert.Equalf(t, "", rx1, "Received a message on handler 1 after unsubscribe")
	assert.Equalf(t, "", rx2, "Did not receive the message on handler 2")
	rxMutex.Unlock()

	gwm.Disconnect()
}

// TMessengerBadUnsubscribe tests unsubscribe of not subscribed channel
func TMessengerBadUnsubscribe(t *testing.T, gwm messaging.IGatewayMessenger) {
	// clientID := "test"
	const timeout = 10

	err := gwm.Connect(clientID, timeout)
	require.NoError(t, err)

	gwm.Unsubscribe(messaging.TestChannelID)
	gwm.Disconnect()

}

// TMessengerPubNoConnect helper to test  messenger connection
func TMessengerPubNoConnect(t *testing.T, gwm messaging.IGatewayMessenger) {
	// clientID := "test"
	const timeout = 10
	var msg1 = "Hello world 1"

	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	err := gwm.Publish(messaging.TestChannelID, []byte(msg1))
	require.Error(t, err)

	gwm.Disconnect()

}

// TMessengerSubBeforeConnect subscribe should work before connection is established
func TMessengerSubBeforeConnect(t *testing.T, gwm messaging.IGatewayMessenger) {
	// clientID := "test"
	const timeout = 10
	const msg = "hello 1"
	var rx string
	rxMutex := sync.Mutex{}
	// mqttMessenger := NewMqttMessenger(clientID, mqttCertFolder)

	handler1 := func(channel string, msg []byte) {
		logrus.Infof("Received message on handler 1: %s", msg)
		rxMutex.Lock()
		defer rxMutex.Unlock()
		rx = string(msg)
	}
	gwm.Subscribe(messaging.TestChannelID, handler1)

	err := gwm.Connect(clientID, timeout)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	err = gwm.Publish(messaging.TestChannelID, []byte(msg))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	rxMutex.Lock()
	assert.Equal(t, msg, rx)
	rxMutex.Unlock()

	gwm.Disconnect()

}
