// Package lib with gateway internal service bus connection functions
// Intended for use by testing and the ISB connection library
package lib

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// const publishAddress = "ws://%s/channel/%s/pub"
// const subscriberAddress = "ws://%s/channel/%s/sub"

// Client and server certificates
const (
	CaCertFile     = "ca.crt"
	CaKeyFile      = "ca.key"
	ServerCertFile = "gateway.crt"
	ServerKeyFile  = "gateway.key"
	ClientCertFile = "client.crt"
	ClientKeyFile  = "client.key"
)

// Connection headers
const (
	AuthorizationHeader = "Authorization"
	ClientHeader        = "Client"
)

//--- communication functions that do the actual work

// newWebsocketConnection creates a new connection for the given path.
// Client certificates are used for authentication and server certificate for server authentication.
// This returns a websocket connection
// If onReceiveHandler returns a result, the result is send as a response to the channel.
// Only use this if a result is expected, otherwise return nil
// clientID is the username of the client that is connecting and included in the header
// clientCertPEM is the client certificate used to verify the client with the server
// clientKeyPEM is the client certificate key used to verify the client with the server
// serverCertPEM is the CA to verify the server with the client
func newWebsocketConnection(url string, clientID string,
	clientCertPEM []byte, clientKeyPEM []byte, serverCertPEM []byte,
	onReceiveHandler func(message []byte, isClosed bool)) (*websocket.Conn, error) {

	// logrus.Infof("NewChannelConnection: connecting to %s with client ID %s", url, clientID)
	socketDialer := websocket.DefaultDialer

	// Use client certificate to identify with the server
	if clientCertPEM != nil {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(serverCertPEM)

		clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
		if err != nil {
			logrus.Error("newWebsocketConnection: Invalid client certificate or key: ", err)
			return nil, err
		}

		socketDialer = &websocket.Dialer{}
		socketDialer.TLSClientConfig = &tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{clientCert},
		}
	}
	reqHeader := http.Header{}
	reqHeader.Add(ClientHeader, clientID)
	// reqHeader.Add(AuthorizationHeader, authToken)

	// Include the client ID using BASIC authentication
	if clientID != "" {
		authToken := "newWebsocketConnection"
		basicAuthField := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientID+":"+authToken))
		// h := http.Header{"Authorization", {"Basic " + base64.StdEncoding.EncodeToString([]byte(username + ":" + password))}}
		reqHeader.Add(AuthorizationHeader, basicAuthField)
	}

	connection, resp, err := socketDialer.Dial(url, reqHeader)
	if err != nil {
		msg := fmt.Sprintf("%s: %s", url, err)
		if resp != nil {
			msg = fmt.Sprintf("%s: %s (%d)", err, resp.Status, resp.StatusCode)
		}
		logrus.Error("newWebsocketConnection: Failed to connect: ", msg)
		return nil, err
	}
	// setup a receive loop for this connection if a receive handler is provided
	// also listen on publisher connections to detect connection closure
	// if onReceiveHandler != nil {
	go func() {
		remoteURL := url
		conn := connection
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				// the connect has closed
				// logrus.Warningf("NewChannelConnection: Connection to %s has closed", url)
				logrus.Warningf("newWebsocketConnection: ReadMessage, connection to %s closed", remoteURL)
				err = conn.Close()
				if onReceiveHandler != nil {
					// FIXME: inform publisher that its connection has closed
					onReceiveHandler(nil, true)
				}
				break
			}
			// logrus.Infof("NewChannelConnection: Received message on %s", url)
			if onReceiveHandler != nil {
				onReceiveHandler(message, false)
			}
		}
	}()
	// }

	return connection, nil
}

// NewPublisher creates a new connection to publish on a channel
// clientID is the ID of the publisher that is connecting
// This returns a websocket connection
func NewPublisher(hostPort string, clientID string, channelID string) (*websocket.Conn, error) {
	const publishAddress = "ws://%s/channel/%s/pub"
	url := fmt.Sprintf(publishAddress, hostPort, channelID)
	return newWebsocketConnection(url, clientID, nil, nil, nil, nil)
}

// NewSubscriber creates a new connection for a subscriber to a channel
// clientID is the ID of the subscriber that is connecting
// handler is invoked when a message is to be processed. It should return the provided or modified message
// This returns a websocket connection
func NewSubscriber(hostPort string, clientID string, channelID string,
	handler func(channel string, msg []byte)) (*websocket.Conn, error) {
	const subscriberAddress = "ws://%s/channel/%s/sub"
	var myChannelID = channelID

	url := fmt.Sprintf(subscriberAddress, hostPort, channelID)
	return newWebsocketConnection(url, clientID, nil, nil, nil, func(msg []byte, isClosed bool) {
		// if isClosed
		handler(myChannelID, msg)
	})
}

// NewTLSPublisher creates a new TLS connection to publish on a channel.
// This uses both a Certificate Authority and Client certificate to verify both client and server to each other.
// clientID is the ID of the publisher that is connecting
// clientCertPEM is the client certificate used to verify the client with the server
// clientKeyPEM is the client certificate key used to verify the client with the server
// serverCertPEM is the CA to verify the server with the client
// This returns a websocket connection
func NewTLSPublisher(hostPort string, clientID string, channelID string,
	clientCertPEM []byte, clientKeyPEM []byte, serverCertPEM []byte) (*websocket.Conn, error) {
	const publishAddress = "wss://%s/channel/%s/pub"
	url := fmt.Sprintf(publishAddress, hostPort, channelID)
	return newWebsocketConnection(url, clientID, clientCertPEM, clientKeyPEM, serverCertPEM, nil)
}

// NewTLSSubscriber creates a new TLS connection for a subscriber to a channel
// clientID is the ID of the subscriber that is connecting
// clientCertPEM is the client certificate used to verify the client with the server
// clientKeyPEM is the client certificate key used to verify the client with the server
// serverCertPEM is the CA to verify the server with the client
// handler is invoked when a message is to be processed. It should return the provided or modified message
// This returns a websocket connection
func NewTLSSubscriber(hostPort string, clientID string, channelID string,
	clientCertPEM []byte, clientKeyPEM []byte, serverCertPEM []byte,
	handler func(channel string, msg []byte)) (*websocket.Conn, error) {

	const subscriberAddress = "wss://%s/channel/%s/sub"
	var myChannelID = channelID

	url := fmt.Sprintf(subscriberAddress, hostPort, channelID)
	return newWebsocketConnection(url, clientID, clientCertPEM, clientKeyPEM, serverCertPEM, func(msg []byte, isClosed bool) {
		handler(myChannelID, msg)
	})
}

// SendMessage over the websocket connection
func SendMessage(conn *websocket.Conn, message []byte) error {
	err := conn.WriteMessage(websocket.TextMessage, message)
	return err
}

// // SendMessage sends a message into the channel
// func SendMessage(connection *websocket.Conn, message []byte) error {
// 	w, err := connection.NextWriter(websocket.TextMessage)
// 	if err != nil {
// 		return err
// 	}
// 	w.Write(message)
// 	w.Close()
// 	return nil
// }
