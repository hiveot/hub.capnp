package plugin

import (
	"io/ioutil"
	"path"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/src/lib"
)

// const publishAddress = "ws://%s/channel/%s/pub"
// const subscriberAddress = "ws://%s/channel/%s/sub"

// Connection headers
const (
	AuthorizationHeader = "Authorization"
	ClientHeader        = "Client"
)

// ISBMessenger is the plugin messenger to the gateway's internal service bus
// This implements the IGatewayConnection interface
type ISBMessenger struct {
	clientID      string                     // Who Am I?
	serverAddress string                     // hostname/ip:port of the server
	clientCertPEM []byte                     // client certificate to authenticate with the server
	clientKeyPEM  []byte                     // private key of this client certificate
	serverCertPEM []byte                     // server certificate to verify the gateway against
	publishers    map[string]*websocket.Conn // channel to publisher connection map
	subscribers   map[string]*websocket.Conn // channel to subscriber connection map
	updateMutex   *sync.Mutex
}

// Connect to the internal service bus server
// This doesn't connect yet until publish or subscribe is called
func (isb *ISBMessenger) Connect(serverAddress string) error {
	isb.serverAddress = serverAddress
	// TBD we could do a connection attempt to validate it
	return nil
}

// Disconnect all connections and stop listeners
func (isb *ISBMessenger) Disconnect() {
	isb.updateMutex.Lock()
	defer isb.updateMutex.Unlock()
	for _, sub := range isb.subscribers {
		logrus.Infof("Close: subscription connection to: %s", sub.RemoteAddr())
		sub.Close()
	}
	for _, pub := range isb.publishers {
		logrus.Infof("Close: publisher connection to: %s", pub.RemoteAddr())
		pub.Close()
	}
	isb.publishers = make(map[string]*websocket.Conn)
	isb.subscribers = make(map[string]*websocket.Conn)
}

// Publish sends a message into a channel
// This creates a new websocket connection if one doesn't yet exist. If writing fails
// the connection is removed. The caller can simply retry to publish which attempts to create
// a new connection.
func (isb *ISBMessenger) Publish(channel string, message []byte) error {
	conn, err := isb.getPublisher(channel)
	if err != nil {
		return err
	}

	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		logrus.Warningf("Publish: writing failed to: %s. Connection removed.", conn.RemoteAddr())
		delete(isb.publishers, channel)
		return err
	}
	_, err = w.Write(message)
	w.Close()
	return err
}

// Subscribe to a channel
// This creates a websocket connection.
// If certificates are available then create the connection over TLS
// The subscription remains for the lifecycle of the plugin
func (isb *ISBMessenger) Subscribe(
	channel string, handler func(channel string, message []byte)) error {
	var conn *websocket.Conn
	var err error
	// FIXME: handle the server dropping the connection
	if isb.clientCertPEM != nil {
		conn, err = lib.NewTLSSubscriber(isb.serverAddress, isb.clientID, channel,
			isb.clientCertPEM, isb.clientKeyPEM, isb.serverCertPEM, handler)
		isb.subscribers[channel] = conn
		return err
	}
	conn, err = lib.NewSubscriber(isb.serverAddress, isb.clientID, channel, handler)
	isb.updateMutex.Lock()
	defer isb.updateMutex.Unlock()
	isb.subscribers[channel] = conn

	return err
}

// getPublisher returns a connection for the given channel.
// If the publisher doesn't exist, then create a new one. Use TLS if certificates are available.
func (isb *ISBMessenger) getPublisher(channel string) (conn *websocket.Conn, err error) {
	conn = isb.publishers[channel]
	if conn == nil {
		if isb.clientCertPEM != nil {
			conn, err = lib.NewTLSPublisher(isb.serverAddress, isb.clientID, channel,
				isb.clientCertPEM, isb.clientKeyPEM, isb.serverCertPEM)
		} else {
			conn, err = lib.NewPublisher(isb.serverAddress, isb.clientID, channel)
		}
		if err != nil {
			logrus.Errorf("getPublisher: connection to server %s on channel %s failed: %s", isb.serverAddress, channel, err)
			return
		}
		isb.publishers[channel] = conn
		conn.SetCloseHandler(func(code int, text string) error {
			isb.updateMutex.Lock()
			defer isb.updateMutex.Unlock()

			delete(isb.publishers, channel)
			return nil
		})
	}
	return
}

// NewISBMessenger creates a new instance of the internal service bus messenger to publish
// and subscribe to gateway messages.
func NewISBMessenger(clientID string, certFolder string) *ISBMessenger {
	isb := &ISBMessenger{
		clientID: clientID,
		// serverAddress: serverAddress,
		publishers:  make(map[string]*websocket.Conn),
		subscribers: make(map[string]*websocket.Conn),
		updateMutex: &sync.Mutex{},
	}
	if certFolder != "" {
		// caCertPath := path.Join(certFolder, CaCertFile)
		// caKeyPath := path.Join(certFolder, CaKeyFile)
		serverCertPath := path.Join(certFolder, ServerCertFile)
		// serverKeyPath := path.Join(certFolder, ServerKeyFile)
		clientCertPath := path.Join(certFolder, ClientCertFile)
		clientKeyPath := path.Join(certFolder, ClientKeyFile)

		isb.serverCertPEM, _ = ioutil.ReadFile(serverCertPath)
		// gwsb.serverKeyPEM, _ := ioutil.ReadFile(serverKeyFile),
		isb.clientCertPEM, _ = ioutil.ReadFile(clientCertPath)
		isb.clientKeyPEM, _ = ioutil.ReadFile(clientKeyPath)

	}
	return isb
}
