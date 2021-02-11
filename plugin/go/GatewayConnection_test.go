package plugin

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/gateway/src/servicebus"
)

const certFolder = "../../test"

// Test create and close the internal service bus channel
func TestISBConnection(t *testing.T) {
	clientID := "test"
	logrus.Info("Testing create channels")
	gwc := NewGatewayConnection(clientID, ConnectionProtocolISB, certFolder)
	gwc.Disconnect()
	// _ = gwc
}

func TestNoClientID(t *testing.T) {
	clientID := ""
	gwc := NewGatewayConnection(clientID, ConnectionProtocolISB, certFolder)
	require.Nil(t, gwc)
}

func TestInvalidProtocol(t *testing.T) {
	clientID := "test"
	gwc := NewGatewayConnection(clientID, "invalid", certFolder)
	require.Nil(t, gwc)
}

func TestConnectMqtt(t *testing.T) {
	clientID := "test"
	gwc := NewGatewayConnection(clientID, ConnectionProtocolMQTT, certFolder)
	require.NotNil(t, gwc)
	// err := gwc.connection.Connect(serverAddr)
	// err := gwc.Publish("test1", nil)
	// assert.Error(t, err, "Publish to invalid server should fail")
	gwc.Disconnect()
}

func TestPubSubISBNoTLS(t *testing.T) {
	var rx string
	var msg1 = "Hello world"
	clientID := "test"
	serverAddr := "localhost:9678" // default

	isbServer, err := servicebus.StartServiceBus(serverAddr)
	_ = isbServer
	require.NoError(t, err, "Failed starting the ISB server")
	time.Sleep(10 * time.Millisecond)

	gwc := NewGatewayConnection(clientID, ConnectionProtocolISB, "")
	err = gwc.Connect(serverAddr)
	require.NoError(t, err)
	err = gwc.Subscribe(TestChannelID, func(channel string, msg []byte) {
		rx = string(msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", TestChannelID)

	err = gwc.Publish(TestChannelID, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")
	time.Sleep(10 * time.Millisecond)
	require.Equalf(t, msg1, rx, "Did not receive the message")

	gwc.Disconnect()
	isbServer.Stop()
}
func TestPubSubISBWithTLS(t *testing.T) {
	var rx string
	var msg1 = "Hello world"
	clientID := "test"
	serverAddr := "localhost:9678" // default
	// certFolder := ""

	isbServer, err := servicebus.StartTLSServiceBus(serverAddr, certFolder)
	_ = isbServer
	require.NoError(t, err, "Failed starting the ISB server")
	time.Sleep(10 * time.Millisecond)

	gwc := NewGatewayConnection(clientID, ConnectionProtocolISB, certFolder)
	err = gwc.Connect(serverAddr)
	require.NoError(t, err)
	err = gwc.Subscribe(TestChannelID, func(channel string, msg []byte) {
		rx = string(msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", TestChannelID)

	err = gwc.Publish(TestChannelID, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")
	time.Sleep(10 * time.Millisecond)
	require.Equalf(t, msg1, rx, "Did not receive the message")

	gwc.Disconnect()
	isbServer.Stop()
}

func TestWriteToClosedConnection(t *testing.T) {
	var rx string
	var msg1 = "Hello world"
	clientID := "test"
	serverAddr := "localhost:9678" // default

	isbServer, err := servicebus.StartServiceBus(serverAddr)
	_ = isbServer
	require.NoError(t, err, "Failed starting the ISB server")
	time.Sleep(10 * time.Millisecond)

	gwc := NewGatewayConnection(clientID, ConnectionProtocolISB, "")
	err = gwc.Connect(serverAddr)
	require.NoError(t, err)
	err = gwc.Subscribe(TestChannelID, func(channel string, msg []byte) {
		rx = string(msg)
	})
	require.NoErrorf(t, err, "Failed subscribing to channel %s", TestChannelID)

	err = gwc.Publish(TestChannelID, []byte(msg1))
	require.NoErrorf(t, err, "Failed publishing message")
	time.Sleep(10 * time.Millisecond)
	assert.Equalf(t, msg1, rx, "Did not receive the message")
	_ = rx

	logrus.Infof("TestWriteToClosedConnection: Stopping server")
	isbServer.Stop()

	// the client connection should close when the server shuts down
	time.Sleep(1000 * time.Millisecond)
	logrus.Infof("TestWriteToClosedConnection: Publishing message")
	// for some reason the first message still succeeds
	err = gwc.Publish(TestChannelID, []byte(msg1))
	// the second message should surely fail
	err = gwc.Publish(TestChannelID, []byte(msg1))
	assert.Errorf(t, err, "Expected error publishing 1st message to closed connection")

	time.Sleep(1000 * time.Millisecond)
	gwc.Disconnect()
	isbServer.Stop()
}
