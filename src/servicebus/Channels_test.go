package servicebus_test

import (
	"testing"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/src/servicebus"
)

const channel1ID = "channel1"
const channel2ID = "channel2"
const defaultBufferSize = 1

// Test create, store and remove channels
func TestCreateChannel(t *testing.T) {
	logrus.Info("Testing create channels")
	cp := servicebus.NewChannelPlumbing()
	c1 := &websocket.Conn{}
	c2 := &websocket.Conn{}
	c3 := &websocket.Conn{}
	channel1 := cp.NewChannel(channel1ID, defaultBufferSize)
	cp.AddSubscriber(channel1, c1)
	cp.AddSubscriber(channel1, c2)
	cp.AddSubscriber(channel1, c3)

	clist1 := cp.GetSubscribers(channel1ID)
	clist2 := cp.GetSubscribers(channel2ID)
	assert.Equal(t, 3, len(clist1), "Expected 3 subscriber in channel 1")
	assert.Equal(t, 0, len(clist2), "Expected 0 subscribers in channel 2")

	removeSuccessful := cp.RemoveConnection(c1)
	assert.True(t, removeSuccessful, "Connection c1 should have been found")
	removeSuccessful = cp.RemoveConnection(c2)
	assert.True(t, removeSuccessful, "Connection c2 should have been found")
	clist1 = cp.GetSubscribers(channel1ID)
	assert.Equal(t, 1, len(clist1), "Expected 1 remaining connection in channel 1")

	// removing twice should not fail
	cp.RemoveConnection(c1)
}
