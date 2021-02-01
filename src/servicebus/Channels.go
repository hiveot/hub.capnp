// Package servicebus for managing channel connections
package servicebus

import (
	"sync"

	"github.com/gorilla/websocket"
)

// ChannelPlumbing contains multiple channels
type ChannelPlumbing struct {
	channels    map[string]*Channel // channel ID
	updateMutex *sync.Mutex
}

// Channel holding publisher and subscriber connections to this channel
type Channel struct {
	ID            string            // ID of the channel
	publishers    []*websocket.Conn // publisher connections
	subscribers   []*websocket.Conn // subscriber connections
	ChannelGoChan chan []byte       // job queue for processing
}

// AddPublisher adds a publisher to a channel
// c contains the connection to add
func (cp *ChannelPlumbing) AddPublisher(channel *Channel, c *websocket.Conn) *Channel {
	cp.updateMutex.Lock()
	defer cp.updateMutex.Unlock()
	channel.publishers = append(channel.publishers, c)
	return channel
}

// AddSubscriber adds a subscriber connection to a channel
// c contains the connection to add
func (cp *ChannelPlumbing) AddSubscriber(channel *Channel, c *websocket.Conn) *Channel {
	cp.updateMutex.Lock()
	defer cp.updateMutex.Unlock()
	channel.subscribers = append(channel.subscribers, c)
	return channel
}

// GetSubscribers returns a list of channel subscriber connections
func (cp *ChannelPlumbing) GetSubscribers(channelID string) []*websocket.Conn {
	cp.updateMutex.Lock()
	defer cp.updateMutex.Unlock()
	channel := cp.channels[channelID]
	if channel == nil {
		return make([]*websocket.Conn, 0)
	}
	return channel.subscribers
}

// GetChannel returns the channel for the given ID
func (cp *ChannelPlumbing) GetChannel(channelID string) *Channel {
	channel := cp.channels[channelID]
	return channel
}

// NewChannel creates a new channel infrastructure for the given ID
func (cp *ChannelPlumbing) NewChannel(channelID string, bufferSize int) *Channel {
	channel := &Channel{
		ID:            channelID,
		publishers:    make([]*websocket.Conn, 0),
		subscribers:   make([]*websocket.Conn, 0),
		ChannelGoChan: make(chan []byte, bufferSize),
	}
	cp.channels[channelID] = channel

	return channel
}

// RemoveConnection a channel connection while retaining order
// Returns true if remove successful, false if connection not found
func (cp *ChannelPlumbing) RemoveConnection(c *websocket.Conn) bool {
	cp.updateMutex.Lock()
	defer cp.updateMutex.Unlock()
	// slow way is okay as there won't be that many connections
	for _, channel := range cp.channels {
		for index, connection := range channel.subscribers {
			if connection == c {
				if index == len(channel.subscribers)-1 {
					channel.subscribers = channel.subscribers[:index]
				} else {
					channel.subscribers = append(channel.subscribers[:index], channel.subscribers[index+1:]...)
				}
				return true
			}
		}
		for index, connection := range channel.publishers {
			if connection == c {
				if index == len(channel.publishers)-1 {
					channel.publishers = channel.publishers[:index]
				} else {
					channel.publishers = append(channel.publishers[:index], channel.publishers[index+1:]...)
				}
				return true
			}
		}
	}
	return false
}

// NewChannelPlumbing creates an instance of channels
func NewChannelPlumbing() ChannelPlumbing {
	p := ChannelPlumbing{}
	p.updateMutex = &sync.Mutex{}
	p.channels = make(map[string]*Channel, 0)
	return p
}
