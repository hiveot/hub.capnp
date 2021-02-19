package smbus

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/certs"
)

// const publishAddress = "ws://%s/channel/%s/pub"
// const subscriberAddress = "ws://%s/channel/%s/sub"

// SmbusMessenger provides the IGatewayMessenger API for the simple message bus
type SmbusMessenger struct {
	clientID      string          // Who Am I?
	hostPort      string          // hostname/ip:port of the server
	clientCertPEM []byte          // client certificate to authenticate with the server
	clientKeyPEM  []byte          // private key of this client certificate
	serverCertPEM []byte          // server certificate to verify the gateway against
	connection    *websocket.Conn // websocket connection to internal service bus
	subscribers   map[string]func(channelID string, msg []byte)
	updateMutex   *sync.Mutex
}

// Connect to the internal message bus server
func (smbmsg *SmbusMessenger) Connect(clientID string, timeoutSec int) error {
	var conn *websocket.Conn
	var err error
	hostName, _ := os.Hostname()
	if clientID == "" {
		clientID = fmt.Sprintf("%s-%d", hostName, time.Now().Unix())
	}
	smbmsg.clientID = clientID
	// TBD we could do a connection attempt to validate it

	if smbmsg.clientCertPEM != nil {
		conn, err = NewTLSWebsocketConnection(
			smbmsg.hostPort, clientID, smbmsg.onReceiveMessage,
			smbmsg.clientCertPEM, smbmsg.clientKeyPEM, smbmsg.serverCertPEM)
	} else {
		conn, err = NewWebsocketConnection(
			smbmsg.hostPort, clientID, smbmsg.onReceiveMessage)
	}
	smbmsg.connection = conn
	// subscribe to existing channels
	for channelID := range smbmsg.subscribers {
		Subscribe(smbmsg.connection, channelID)
	}

	return err
}

// Disconnect all connections and stop listeners
func (smbmsg *SmbusMessenger) Disconnect() {
	smbmsg.updateMutex.Lock()
	defer smbmsg.updateMutex.Unlock()

	if smbmsg.connection != nil {
		smbmsg.connection.Close()
		smbmsg.connection = nil
	}
}

// Receive a subscribed message and pass it to its handler
func (smbmsg *SmbusMessenger) onReceiveMessage(command string, channelID string, message []byte) {
	logrus.Infof("onReceiveMessage: command=%s, channelID=%s", command, channelID)
	if command == MsgBusCommandReceive {
		handler := smbmsg.subscribers[channelID]
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
func (smbmsg *SmbusMessenger) Publish(channelID string, message []byte) error {
	return Publish(smbmsg.connection, channelID, message)
}

// Subscribe to a channel. Existing subscriptions are replaced
// wildcards are not supported
func (smbmsg *SmbusMessenger) Subscribe(
	channelID string, handler func(channel string, message []byte)) {

	smbmsg.updateMutex.Lock()
	defer smbmsg.updateMutex.Unlock()
	// remove any previous subscriptions
	smbmsg.subscribers[channelID] = handler
	Subscribe(smbmsg.connection, channelID)
}

// Unsubscribe from a channel
func (smbmsg *SmbusMessenger) Unsubscribe(channelID string) {
	smbmsg.updateMutex.Lock()
	defer smbmsg.updateMutex.Unlock()
	smbmsg.subscribers[channelID] = nil
	Unsubscribe(smbmsg.connection, channelID)
}

// NewSmbusMessenger creates a new instance of the lightweigh websocket messenger to publish
// and subscribe to gateway messages.
func NewSmbusMessenger(certFolder string, hostPort string) *SmbusMessenger {

	smbmsg := &SmbusMessenger{
		// serverAddress: serverAddress,
		subscribers: make(map[string]func(channelID string, msg []byte)),
		updateMutex: &sync.Mutex{},
		hostPort:    hostPort,
	}
	if certFolder != "" {
		// caCertPath := path.Join(certFolder, CaCertFile)
		// caKeyPath := path.Join(certFolder, CaKeyFile)
		serverCertPath := path.Join(certFolder, certs.ServerCertFile)
		// serverKeyPath := path.Join(certFolder, certs.ServerKeyFile)
		clientCertPath := path.Join(certFolder, certs.ClientCertFile)
		clientKeyPath := path.Join(certFolder, certs.ClientKeyFile)

		smbmsg.serverCertPEM, _ = ioutil.ReadFile(serverCertPath)
		// gwsb.serverKeyPEM, _ := ioutil.ReadFile(serverKeyFile),
		smbmsg.clientCertPEM, _ = ioutil.ReadFile(clientCertPath)
		smbmsg.clientKeyPEM, _ = ioutil.ReadFile(clientKeyPath)
		if smbmsg.serverCertPEM == nil {
			logrus.Errorf("NewSmbusMessenger: no certificates in %s", certFolder)
		}
	}
	return smbmsg
}
