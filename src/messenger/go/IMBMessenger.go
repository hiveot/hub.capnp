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
	"github.com/wostzone/gateway/src/msgbus"
)

// const publishAddress = "ws://%s/channel/%s/pub"
// const subscriberAddress = "ws://%s/channel/%s/sub"

// Connection headers
const (
	AuthorizationHeader = "Authorization"
	ClientHeader        = "Client"
)

// IMBMessenger is the internal message bus for plugin to gateway communication
// This implements the IGatewayConnection interface
type IMBMessenger struct {
	clientID      string          // Who Am I?
	serverAddress string          // hostname/ip:port of the server
	clientCertPEM []byte          // client certificate to authenticate with the server
	clientKeyPEM  []byte          // private key of this client certificate
	serverCertPEM []byte          // server certificate to verify the gateway against
	connection    *websocket.Conn // websocket connection to internal service bus
	subscribers   map[string]func(channelID string, msg []byte)
	updateMutex   *sync.Mutex
}

// Connect to the internal service bus server
func (isb *IMBMessenger) Connect(serverAddress string, clientID string, timeoutSec int) error {
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
		conn, err = msgbus.NewTLSWebsocketConnection(
			isb.serverAddress, clientID, isb.onReceiveMessage,
			isb.clientCertPEM, isb.clientKeyPEM, isb.serverCertPEM)
	} else {
		conn, err = msgbus.NewWebsocketConnection(
			isb.serverAddress, clientID, isb.onReceiveMessage)
	}
	isb.connection = conn
	// subscribe to existing channels
	for channelID, _ := range isb.subscribers {
		msgbus.Subscribe(isb.connection, channelID)
	}

	return err
}

// Disconnect all connections and stop listeners
func (isb *IMBMessenger) Disconnect() {
	isb.updateMutex.Lock()
	defer isb.updateMutex.Unlock()

	if isb.connection != nil {
		isb.connection.Close()
		isb.connection = nil
	}
}

// Receive a subscribed message and pass it to its handler
func (isb *IMBMessenger) onReceiveMessage(command string, channelID string, message []byte) {
	logrus.Infof("onReceiveMessage: command=%s, channelID=%s", command, channelID)
	if command == msgbus.MsgBusCommandReceive {
		handler := isb.subscribers[channelID]
		if handler == nil {
			logrus.Errorf("onReceiveMessage: Missing handler for channel %s. Message ignored.", channelID)
		} else {
			handler(channelID, message)
		}
	} else {
		logrus.Warningf("onReceiveMessage: Unexpected command %s on channel %s", command, channelID)
	}
}

// Publish sends a message into a channel
// This returns an error if a connection doesn't exist and the message is not delivered
func (isb *IMBMessenger) Publish(channelID string, message []byte) error {
	return msgbus.Publish(isb.connection, channelID, message)
}

// Subscribe to a channel. Existing subscriptions are replaced
// wildcards are not supported
func (isb *IMBMessenger) Subscribe(
	channelID string, handler func(channel string, message []byte)) {

	isb.updateMutex.Lock()
	defer isb.updateMutex.Unlock()
	// remove any previous subscriptions
	isb.subscribers[channelID] = handler
	msgbus.Subscribe(isb.connection, channelID)
}

// Unsubscribe from a channel
func (isb *IMBMessenger) Unsubscribe(channelID string) {
	isb.updateMutex.Lock()
	defer isb.updateMutex.Unlock()
	isb.subscribers[channelID] = nil
	msgbus.Unsubscribe(isb.connection, channelID)
}

// NewISBMessenger creates a new instance of the internal service bus messenger to publish
// and subscribe to gateway messages.
func NewISBMessenger(certFolder string) *IMBMessenger {

	isb := &IMBMessenger{
		// serverAddress: serverAddress,
		subscribers: make(map[string]func(channelID string, msg []byte)),
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
