// Package smbus with client side helper functions to connect, send and receive messages
package smbus

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// messagebus server address composition
const (
	MsgbusAddress = "/wost"
	MsgbusURL     = "ws://%s" + MsgbusAddress
	MsgbusTLSURL  = "wss://%s" + MsgbusAddress
)

// Built-in message bus commands
const (
	MsgBusCommandPublish     = "publish" // publish by publisher
	MsgBusCommandReceive     = "receive" // receive by subscriber
	MsgBusCommandSubscribe   = "subscribe"
	MsgBusCommandUnsubscribe = "unsubscribe"
)

// Connection headers
const (
	AuthorizationHeader = "Authorization"
	ClientHeader        = "Client"
)

// DefaultSmbusHost with listening address and port
const DefaultSmbusHost = "localhost:9678"

//--- communication functions that do the actual work

// Connect to the simple message bus server
//  socketdialer is used to connect and is setup with TLS certificates if applicable
//  url contains the full websocket URL, eg ws://host:port/path or wss://host:port/path
//  clientID is used to identify this client
func Connect(socketDialer *websocket.Dialer, url string, clientID string) (*websocket.Conn, error) {
	reqHeader := http.Header{}
	reqHeader.Add(ClientHeader, clientID)

	connection, resp, err := socketDialer.Dial(url, reqHeader)
	if err != nil {
		msg := fmt.Sprintf("%s: %s", url, err)
		if resp != nil {
			msg = fmt.Sprintf("%s: %s (%d)", err, resp.Status, resp.StatusCode)
		}
		logrus.Error("connect: Failed to connect: ", msg)
		return nil, err
	}
	return connection, err
}

// Listen for data from the connection and determine the command, channel and the message
// This function blocks until the connection is closed
//  conn is an active connection to a websocket server
//  handler is a message callback
func Listen(conn *websocket.Conn, handler func(command string, topic string, message []byte)) {
	// setup a receive loop for this connection if a receive handler is provided
	// also listen on publisher connections to detect connection closure
	remoteURL := conn.RemoteAddr()
	// conn := connection
	for {
		msgType, message, err := conn.ReadMessage()
		_ = msgType
		if err != nil {
			// the connect has closed
			// logrus.Warningf("NewChannelConnection: Connection to %s has closed", url)
			logrus.Warningf("listen: ReadMessage, read error from %s: %s", remoteURL, err)
			err = conn.Close()
			if handler != nil {
				handler("", "", nil)
			}
			break
		}
		// message contains command:topic:data
		command := ""
		topic := ""
		data := []byte(nil)
		parts := strings.SplitN(string(message), ":", 3)
		if len(parts) != 3 {
			logrus.Warningf("listen: Ignored invalid message without command or topic from %s", remoteURL)
		} else {
			command = parts[0]
			topic = parts[1]
			data = []byte(parts[2])
			if handler != nil {
				handler(command, topic, data)
			}
		}
	}
}

// Send a command to the message bus server
func Send(conn *websocket.Conn, command string, channelID string, message []byte) error {
	if conn == nil {
		msg := fmt.Errorf("send: Unable to deliver command %s to channel %s. No connection to server", command, channelID)
		return msg
	}

	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		logrus.Errorf("send: writing failed to: %s. Connection broken.", conn.RemoteAddr())
		// should we retry?
		return err
	}
	// message is simply encoded with command:topic:message
	payload := fmt.Sprintf("%s:%s:%s", command, channelID, message)

	_, err = w.Write([]byte(payload))
	w.Close()
	return err
}

// NewWebsocketConnection creates a new connection to the websocket server
// clientID is the ID of the publisher that is connecting
// handler is the callback invoked with new messages. Topic is "" and msg is nil when the connection has closed
// This returns a websocket connection
func NewWebsocketConnection(hostPort string, clientID string,
	handler func(command string, channel string, msg []byte)) (*websocket.Conn, error) {

	url := fmt.Sprintf("ws://%s%s", hostPort, MsgbusAddress)

	// logrus.Infof("NewChannelConnection: connecting to %s with client ID %s", url, clientID)
	socketDialer := websocket.DefaultDialer
	logrus.Infof("NewWebsocketConnection: ClientID '%s' connecting to MsgBus URL: %s", clientID, url)
	connection, err := Connect(socketDialer, url, clientID)
	if err == nil {
		// setup a receive loop for this connection if a receive handler is provided
		go func() {
			Listen(connection, handler)
		}()
	}
	return connection, err
}

// NewTLSWebsocketConnection creates a new TLS connection to the websocket server
// This uses both a Certificate Authority and Client certificate to verify both client and server to each other.
// clientID is the ID of the publisher that is connecting
// handler is the callback invoked with new messages. Topic is "" and msg is nil when the connection has closed
// clientCertPEM is the client certificate used to verify the client with the server
// clientKeyPEM is the client certificate key used to verify the client with the server
// serverCertPEM is the CA to verify the server with the client
// This returns a websocket connection
func NewTLSWebsocketConnection(hostPort string, clientID string, handler func(comand string, topic string, msg []byte),
	clientCertPEM []byte, clientKeyPEM []byte, serverCertPEM []byte) (*websocket.Conn, error) {

	url := fmt.Sprintf("wss://%s/wost", hostPort)

	// Use client certificate to identify with the server
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(serverCertPEM)

	clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
	if err != nil {
		logrus.Error("NewTLSWebsocketConnection: Invalid client certificate or key: ", err)
		return nil, err
	}

	socketDialer := &websocket.Dialer{}
	socketDialer.TLSClientConfig = &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{clientCert},
	}

	connection, err := Connect(socketDialer, url, clientID)
	if err == nil {
		// setup a receive loop for this connection if a receive handler is provided
		go func() {
			Listen(connection, handler)
		}()
	}
	return connection, err
}

// Publish sends a publish message into a channel
// This returns an error if a connection doesn't exist and the message is not delivered
func Publish(conn *websocket.Conn, channelID string, message []byte) error {
	return Send(conn, MsgBusCommandPublish, channelID, message)
}

// Subscribe subscribes the connection to a channel
// This returns an error if a connection is closed
func Subscribe(conn *websocket.Conn, channelID string) error {
	return Send(conn, MsgBusCommandSubscribe, channelID, nil)
}

// Unsubscribe removes prior subscription to a channel
// This returns an error if a connection is closed
func Unsubscribe(conn *websocket.Conn, channelID string) error {
	return Send(conn, MsgBusCommandUnsubscribe, channelID, nil)
}
