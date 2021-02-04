package client

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const publishAddress = "ws://%s/channel/%s/pub"
const subscriberAddress = "ws://%s/channel/%s/sub"

// Connection headers
const (
	AuthorizationHeader = "Authorization"
	ClientHeader        = "Client"
)

// newChannelConnection creates a new connection for the given path and authenticates with the
// server using BASIC authentication.
// This returns a websocket connection
// If onReceiveHandler returns a result, the result is send as a response to the channel.
// Only use this if a result is expected, otherwise return nil
// clientID is the username of the client that is connecting
// authToken is the password used to connect
func newChannelConnection(url string, clientID string, authToken string,
	onReceiveHandler func(message []byte, isClosed bool)) (*websocket.Conn, error) {

	// logrus.Infof("NewChannelConnection: connecting to %s with client ID %s", url, clientID)
	reqHeader := http.Header{}
	// reqHeader.Add(AuthorizationHeader, authToken)
	// reqHeader.Add(ClientHeader, clientID)

	if clientID != "" {
		basicAuthField := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+authToken))
		// h := http.Header{"Authorization", {"Basic " + base64.StdEncoding.EncodeToString([]byte(username + ":" + password))}}
		reqHeader.Add(AuthorizationHeader, basicAuthField)
	}
	connection, resp, err := websocket.DefaultDialer.Dial(url, reqHeader)
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
				// logrus.Warningf("NewChannelConnection: Connection to %s has closed", url)
				// onReceiveHandler("", true)
				break
			}
			// logrus.Infof("NewChannelConnection: Received message on %s", url)
			if onReceiveHandler != nil {
				onReceiveHandler(message, false)
			}
		}
	}()

	return connection, nil
}

// NewPublisher creates a new connection to publish on a channel
// clientID is the ID of the publisher that is connecting
// authToken is the authentication token provided to the plugin on startup
// This returns a websocket connection
func NewPublisher(host string, clientID string, authToken string, channelID string) (*websocket.Conn, error) {
	url := fmt.Sprintf(publishAddress, host, channelID)
	return newChannelConnection(url, clientID, authToken, nil)
}

// NewSubscriber creates a new connection for a subscriber to a channel
// clientID is the ID of the subscriber that is connecting
// authToken is the authentication token provided to the plugin on startup
// handler is invoked when a message is to be processed. It should return the provided or modified message
// This returns a websocket connection
func NewSubscriber(host string, clientID string, authToken string, channelID string,
	handler func(msg []byte)) (*websocket.Conn, error) {

	url := fmt.Sprintf(subscriberAddress, host, channelID)
	return newChannelConnection(url, clientID, authToken, func(msg []byte, isClosed bool) {
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
