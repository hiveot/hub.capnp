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
	c11 := &websocket.Conn{}
	c12a := &websocket.Conn{}
	c12b := &websocket.Conn{}
	channel1 := cp.NewChannel(channel1ID, defaultBufferSize)
	cp.AddSubscriber(channel1, c11)
	cp.AddSubscriber(channel1, c12a)
	cp.AddSubscriber(channel1, c12b)

	cl11 := cp.GetSubscribers(channel1ID)
	cl12 := cp.GetSubscribers(channel1ID)
	cl21 := cp.GetSubscribers(channel2ID)
	assert.Equal(t, 1, len(cl11), "Expected 1 connection in channel 1")
	assert.Equal(t, 2, len(cl12), "Expected 2 connection in channel 1")
	assert.Equal(t, 0, len(cl21), "Expected 0 connections in channel 2")

	removeSuccessful := cp.RemoveConnection(c11)
	assert.True(t, removeSuccessful, "Connection c11 should have been found")
	removeSuccessful = cp.RemoveConnection(c12a)
	assert.True(t, removeSuccessful, "Connection c12a should have been found")
	cl11 = cp.GetSubscribers(channel1ID)
	cl12 = cp.GetSubscribers(channel1ID)
	assert.Equal(t, 0, len(cl11), "Expected 0 connections in channel")
	assert.Equal(t, 1, len(cl12), "Expected 1 connection in channel")

	// removing twice should not fail
	cp.RemoveConnection(c11)
}
