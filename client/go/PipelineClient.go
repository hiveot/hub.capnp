package client

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const publishAddress = "ws://%s/channel/%s/pub"
const subscriberAddress = "ws://%s/channel/%s/sub"

// NewChannelConnection creates a new connection for the given path and authenticates with the server
// This returns a websocket connection
// If onReceiveHandler returns a result, the result is send as a response to the channel.
// Only use this if a result is expected, otherwise return nil
func NewChannelConnection(url string, authToken string,
	onReceiveHandler func(message []byte, isClosed bool)) (*websocket.Conn, error) {

	logrus.Infof("NewChannelConnection: connecting to %s", url)
	connection, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		msg := fmt.Sprintf("%s: %s", url, err)
		if resp != nil {
			msg = fmt.Sprintf("%s: %s (%d)", err, resp.Status, resp.StatusCode)
		}
		logrus.Error("NewChannelConnection: Failed to connect: ", msg)
		return nil, err
	}
	// setup a receive loop for this client
	go func() {
		for {
			_, message, err := connection.ReadMessage()
			if err != nil {
				// the connect has closed
				logrus.Warningf("NewChannelConnection: Connection to %s has closed", url)
				// onReceiveHandler("", true)
				break
			}
			logrus.Infof("NewChannelConnection: Received message on %s", url)
			if onReceiveHandler != nil {
				onReceiveHandler(message, false)
			}
		}
	}()

	return connection, nil
}

// NewPublisher creates a new connection to publish on a channel
// authToken is the authentication token provided to the plugin on startup
// This returns a websocket connection
func NewPublisher(host string, authToken string, channelID string) (*websocket.Conn, error) {
	url := fmt.Sprintf(publishAddress, host, channelID)
	return NewChannelConnection(url, authToken, nil)
}

// NewSubscriber creates a new connection for a subscriber to a channel
// authToken is the authentication token provided to the plugin on startup
// handler is invoked when a message is to be processed. It should return the provided or modified message
// This returns a websocket connection
func NewSubscriber(host string, authToken string, channelID string,
	handler func(msg []byte)) (*websocket.Conn, error) {

	url := fmt.Sprintf(subscriberAddress, host, channelID)
	return NewChannelConnection(url, authToken, func(msg []byte, isClosed bool) {
		handler(msg)
	})
}

// NewConsumerClient creates a new connection for a channel consumer
// authToken is the authentication token provided to the plugin on startup
// handler is invoked when a message is to be consumed.
// This returns a websocket connection
func NewConsumerClient(host string, authToken string, channelID string,
	handler func(msg []byte)) (*websocket.Conn, error) {

	url := fmt.Sprintf(subscriberAddress, host, channelID)
	return NewChannelConnection(url, authToken, func(msg []byte, isClosed bool) {
		handler(msg)
	})
}

// SendMessage sends a message into the channel
func SendMessage(connection *websocket.Conn, message []byte) error {
	w, err := connection.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	w.Write(message)
	w.Close()
	return nil
}
