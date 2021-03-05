package smbclient

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
	"github.com/wostzone/gateway/pkg/config"
)

// const publishAddress = "ws://%s/channel/%s/pub"
// const subscriberAddress = "ws://%s/channel/%s/sub"

// SmbClient provides the IGatewayMessenger API for the simple message bus
type SmbClient struct {
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
func (smbc *SmbClient) Connect(clientID string, timeoutSec int) error {
	var conn *websocket.Conn
	var err error
	hostName, _ := os.Hostname()
	if clientID == "" {
		clientID = fmt.Sprintf("%s-%d", hostName, time.Now().Unix())
	}
	smbc.clientID = clientID
	// TBD we could do a connection attempt to validate it

	retryDelaySec := 1
	retryDuration := 0
	for timeoutSec == 0 || retryDuration < timeoutSec {

		if smbc.clientCertPEM != nil {
			conn, err = NewTLSWebsocketConnection(
				smbc.hostPort, clientID, smbc.onReceiveMessage,
				smbc.clientCertPEM, smbc.clientKeyPEM, smbc.serverCertPEM)
		} else {
			conn, err = NewWebsocketConnection(
				smbc.hostPort, clientID, smbc.onReceiveMessage)
		}
		if err == nil {
			smbc.connection = conn
			// subscribe to existing channels
			for channelID := range smbc.subscribers {
				Subscribe(smbc.connection, channelID)
			}
			break
		}
		sleepDuration := time.Duration(retryDelaySec)
		retryDuration += int(sleepDuration)
		time.Sleep(sleepDuration * time.Second)
		// slowly increment wait time
		if retryDelaySec < 120 {
			retryDelaySec++
		}
	}
	return err
}

// Disconnect all connections and stop listeners
func (smbc *SmbClient) Disconnect() {
	smbc.updateMutex.Lock()
	defer smbc.updateMutex.Unlock()

	if smbc.connection != nil {
		smbc.connection.Close()
		smbc.connection = nil
	}
}

// Receive a subscribed message and pass it to its handler
func (smbc *SmbClient) onReceiveMessage(command string, channelID string, message []byte) {
	logrus.Infof("onReceiveMessage: command=%s, channelID=%s", command, channelID)
	if command == MsgBusCommandReceive {
		smbc.updateMutex.Lock()
		handler := smbc.subscribers[channelID]
		defer smbc.updateMutex.Unlock()
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
func (smbc *SmbClient) Publish(channelID string, message []byte) error {
	return Publish(smbc.connection, channelID, message)
}

// Subscribe to a channel. Existing subscriptions are replaced
// wildcards are not supported
func (smbc *SmbClient) Subscribe(
	channelID string, handler func(channel string, message []byte)) {

	smbc.updateMutex.Lock()
	// remove any previous subscriptions
	smbc.subscribers[channelID] = handler
	defer smbc.updateMutex.Unlock()
	Subscribe(smbc.connection, channelID)
}

// Unsubscribe from a channel
func (smbc *SmbClient) Unsubscribe(channelID string) {
	smbc.updateMutex.Lock()
	smbc.subscribers[channelID] = nil
	defer smbc.updateMutex.Unlock()
	Unsubscribe(smbc.connection, channelID)
}

// NewSmbClient creates a new instance of the lightweight messagebus to publish
// and subscribe to gateway messages.
func NewSmbClient(certFolder string, hostPort string) *SmbClient {
	if hostPort == "" {
		hostPort = config.DefaultSmbHost
	}
	smbmsg := &SmbClient{
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
