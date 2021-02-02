package servicebus_test

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	client "github.com/wostzone/gateway/client/go"
	"github.com/wostzone/gateway/src/servicebus"
)

const channel1ID = "channel1"
const channel2ID = "channel2"
const defaultBufferSize = 1

const host = "localhost:9678"

const client1ID = "plugin1"
const authToken1 = "token1"

var authTokens = map[string]string{
	client1ID: authToken1,
}

// Test create, store and remove channels by the server
func TestCreateChannel(t *testing.T) {
	logrus.Info("Testing create channels")
	srv := servicebus.NewChannelServer()
	c1 := &websocket.Conn{}
	c2 := &websocket.Conn{}
	c3 := &websocket.Conn{}
	c4 := &websocket.Conn{}
	channel1 := srv.NewChannel(channel1ID, defaultBufferSize)
	srv.AddSubscriber(channel1, c1)
	srv.AddSubscriber(channel1, c2)
	srv.AddPublisher(channel1, c3)
	srv.AddPublisher(channel1, c4)

	clist1 := srv.GetSubscribers(channel1ID)
	clist2 := srv.GetSubscribers(channel2ID)
	clist3 := srv.GetPublishers(channel1ID)
	clist4 := srv.GetPublishers("not-a-channel")
	assert.Equal(t, 2, len(clist1), "Expected 2 subscriber in channel 1")
	assert.Equal(t, 0, len(clist2), "Expected 0 subscribers in channel 2")
	assert.Equal(t, 2, len(clist3), "Expected 2 publisher in channel 1")
	assert.Equal(t, 0, len(clist4), "Expected 0 publishers in not-a-channel ")

	removeSuccessful := srv.RemoveConnection(c1)
	assert.True(t, removeSuccessful, "Connection c1 should have been found")
	removeSuccessful = srv.RemoveConnection(c2)
	assert.True(t, removeSuccessful, "Connection c2 should have been found")
	removeSuccessful = srv.RemoveConnection(c3)
	assert.True(t, removeSuccessful, "Connection c3 should have been found")
	removeSuccessful = srv.RemoveConnection(c4)
	assert.True(t, removeSuccessful, "Connection c4 should have been found")

	clist1 = srv.GetSubscribers(channel1ID)
	assert.Equal(t, 0, len(clist1), "Expected 0 remaining connections in channel 1")

	// removing twice should not fail
	srv.RemoveConnection(c1)
	srv.RemoveConnection(c4)
}

func TestInvalidAuthentication(t *testing.T) {
	const channel1 = "Chan1"
	const invalidAuthToken = "invalid-token"

	logrus.Infof("Testing authentication on channel %s", channel1)
	cs := servicebus.StartServiceBus(host, authTokens)
	time.Sleep(time.Second)

	_, err1 := client.NewPublisher(host, client1ID, invalidAuthToken, channel1)
	_, err2 := client.NewSubscriber(host, client1ID, invalidAuthToken, channel1, func(msg []byte) {})
	_, err3 := client.NewPublisher(host, "", "", channel1)

	require.Error(t, err1, "Expected error creating publisher with invalid auth")
	require.Error(t, err2, "Expected error creating subscriber with invalid auth")
	require.Error(t, err3, "Expected error creating subscriber with invalid auth")

	cs.Stop()
}

// Test publish and subscribe client
func TestPubSubChannel(t *testing.T) {
	const channel1 = "Chan1"
	const pubMsg1 = "Message 1"
	var subMsg1 = ""

	logrus.Infof("Testing channel %s", channel1)
	cs := servicebus.StartServiceBus(host, authTokens)
	time.Sleep(time.Second)

	// send published channel messages to subscribers
	publisher, err := client.NewPublisher(host, client1ID, authToken1, channel1)
	require.NoError(t, err)

	subscriber, err := client.NewSubscriber(host, client1ID, authToken1, channel1,
		func(msg []byte) {
			logrus.Info("TestChannel: Received published message")
			subMsg1 = string(msg)
		})
	require.NoError(t, err)

	client.SendMessage(publisher, []byte(pubMsg1))
	time.Sleep(1 * time.Second)
	assert.Equal(t, pubMsg1, subMsg1)

	time.Sleep(time.Second * 1)

	// publisher.Close()
	subscriber.Close()
	// time.Sleep(time.Second)
	cs.Stop()
	cs.Stop()

}
