package servicebus_test

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	client "github.com/wostzone/gateway/client/go"
	"github.com/wostzone/gateway/src/servicebus"
)

const host = "localhost:9678"

const authToken1 = "token1"

var authTokens = map[string]string{
	"plugin1": authToken1,
}

// Test pinging the server
func TestPing(t *testing.T) {
	logrus.Info("Testing Ping")
	servicebus.StartServiceBus(host, authTokens)
	// time.Sleep(time.Second)

	t1 := time.Now()
	for i := 0; i < 10; i++ {
		success := client.Ping(host)
		assert.True(t, success, "Failed connecting to server")
	}
	duration := time.Now().Sub(t1) / 10
	logrus.Printf("Duration: %s per ping", duration)

	// time.Sleep(time.Second * 3)
}

// Test publish and subscribe client
func TestChannel(t *testing.T) {
	const channel1 = "Chan1"
	const pubMsg1 = "Message 1"
	var subMsg1 = ""

	logrus.Infof("Testing channel %s", channel1)
	servicebus.StartServiceBus(host, authTokens)
	time.Sleep(time.Second)

	// send published channel messages to subscribers
	publisher, err := client.NewPublisher(host, authToken1, channel1)
	require.NoError(t, err)

	subscriber, err := client.NewSubscriber(host, authToken1, channel1, func(msg []byte) {
		logrus.Info("TestChannel: Received published message")
		subMsg1 = string(msg)
	})
	require.NoError(t, err)

	client.SendMessage(publisher, []byte(pubMsg1))
	time.Sleep(1 * time.Second)
	assert.Equal(t, pubMsg1, subMsg1)

	time.Sleep(time.Second * 1)

	publisher.Close()
	subscriber.Close()
	// time.Sleep(time.Second)
}

// Test pub/sub client authentication
func TestAuthentication(t *testing.T) {
	const channel1 = "Chan1"
	const invalidAuthToken = "invalid-token"

	logrus.Infof("Testing authentication on channel %s", channel1)
	servicebus.StartServiceBus(host, authTokens)
	time.Sleep(time.Second)

	_, err1 := client.NewPublisher(host, invalidAuthToken, channel1)
	_, err2 := client.NewSubscriber(host, invalidAuthToken, channel1, func(msg []byte) {})

	require.Error(t, err1, "Expected error creating publisher with invalid auth")
	require.Error(t, err2, "Expected error creating subscriber with invalid auth")
}
