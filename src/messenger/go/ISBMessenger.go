package messenger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

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
	clientID      string          // Who Am I?
	serverAddress string          // hostname/ip:port of the server
	clientCertPEM []byte          // client certificate to authenticate with the server
	clientKeyPEM  []byte          // private key of this client certificate
	serverCertPEM []byte          // server certificate to verify the gateway against
	connection    *websocket.Conn // websocket connection to internal service bus
	subscribers   map[string]func(topic string, msg []byte)
	updateMutex   *sync.Mutex
}

// Connect to the internal service bus server
func (isb *ISBMessenger) Connect(serverAddress string, clientID string, timeoutSec int) error {
	var conn *websocket.Conn
	var err error
	hostName, _ := os.Hostname()
	if clientID == "" {
		clientID = fmt.Sprintf("%s-%d", hostName, time.Now().Unix())
	}
	isb.clientID = clientID
	isb.serverAddress = serverAddress
	// TBD we could do a connection attempt to validate it

	if isb.clientCertPEM != nil {
		conn, err = lib.NewTLSConnection(isb.serverAddress, clientID, isb.clientCertPEM, isb.clientKeyPEM, isb.serverCertPEM)
	} else {
		conn, err = lib.NewConnection(isb.serverAddress, clientID)
	}
	isb.connection = conn
	return err
}

// Disconnect all connections and stop listeners
func (isb *ISBMessenger) Disconnect() {
	isb.updateMutex.Lock()
	defer isb.updateMutex.Unlock()

	if isb.connection != nil {
		isb.connection.Close()
		isb.connection = nil
	}
}

// Publish sends a message into a channel
// This returns an error if a connection doesn't exist and the message is not delivered
func (isb *ISBMessenger) Publish(topic string, message []byte) error {

	if isb.connection == nil {
		msg := fmt.Errorf("Publish: Unable to deliver message to topic %s. No connection to server", topic)
		return msg
	}

	w, err := isb.connection.NextWriter(websocket.TextMessage)
	if err != nil {
		logrus.Errorf("Publish: writing failed to: %s. Connection broken.", isb.connection.RemoteAddr())
		// should we retry?
		isb.connection = nil
		return err
	}
	// message is simply encoded with topic:message
	topicMessage := topic + ":" + string(message)

	_, err = w.Write([]byte(topicMessage))
	w.Close()
	return err
}

// Subscribe to a channel. Existing subscriptions are replaced
// wildcards are not supported
func (isb *ISBMessenger) Subscribe(
	channel string, handler func(channel string, message []byte)) {

	isb.updateMutex.Lock()
	defer isb.updateMutex.Unlock()
	// remove any previous subscriptions
	isb.subscribers[channel] = handler
}

// Unsubscribe from a channel
func (isb *ISBMessenger) Unsubscribe(channel string) {
	isb.updateMutex.Lock()
	defer isb.updateMutex.Unlock()
	isb.subscribers[channel] = nil
}

// NewISBMessenger creates a new instance of the internal service bus messenger to publish
// and subscribe to gateway messages.
func NewISBMessenger(certFolder string) *ISBMessenger {

	isb := &ISBMessenger{
		// serverAddress: serverAddress,
		subscribers: make(map[string]func(topic string, msg []byte)),
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
