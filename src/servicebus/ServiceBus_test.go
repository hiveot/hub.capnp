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

// Test pinging the server
func TestPing(t *testing.T) {
	logrus.Info("Testing Ping")
	servicebus.StartServiceBus(host)
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

// Test publisher plugin
func TestChannel(t *testing.T) {
	const channel1 = "P1"
	const authToken = ""
	const msg1 = "Message 1"
	var processMsg = ""
	var consumeMsg = ""

	logrus.Infof("Testing channel %s", channel1)
	servicebus.StartServiceBus(host)
	time.Sleep(time.Second)

	// send published channel messages to subscribers
	publisher, err := client.NewPublisher(host, authToken, channel1)
	require.NoError(t, err)

	subscriber, err := client.NewSubscriber(host, authToken, channel1, func(msg []byte) {
		logrus.Info("TestChannel: Received published message")
		consumeMsg = string(msg)
	})
	require.NoError(t, err)

	client.SendMessage(publisher, []byte(msg1))
	time.Sleep(1 * time.Second)
	assert.Equal(t, msg1, processMsg)
	assert.Equal(t, msg1, consumeMsg)

	time.Sleep(time.Second * 1)

	publisher.Close()
	subscriber.Close()
	// time.Sleep(time.Second)

}

// Test capture plugin
func TestCapture(t *testing.T) {

}
