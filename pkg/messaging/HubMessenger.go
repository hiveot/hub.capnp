package messaging

import (
	"github.com/wostzone/hub/pkg/config"
	"github.com/wostzone/hub/pkg/messaging/mqtt"
	"github.com/wostzone/hub/pkg/messaging/smbclient"
)

// ConnectionProtocol of the hub
// type ConnectionProtocol string

// Available message bus connection protocols
const (
	ConnectionProtocolSmbus string = "smbus"
	ConnectionProtocolMQTT  string = "mqtt"
)

// NewHubConnection creates a messenger to a WoST hub message bus.
//
// TLS is disabled if certFolder is empty. smbus generates both server and client certificates if
// the folder is empty. They are used for two-way authentication.
// If the given protocol is unknown, smbus will be used.
//
//  protocol selects the connection method used by the hub: "mqtt" or "smbus".
//  certFolder is the folder containing the client and server certificates for TLS connections
//  hostPort is the hostname and port of the messaging server to connect to
// Returns nil if the connection method is unknown
func NewHubConnection(protocol string, certFolder string, hostPort string) IHubMessenger {
	if protocol == ConnectionProtocolMQTT {
		return mqtt.NewMqttMessenger(certFolder, hostPort)
	}
	return smbclient.NewSmbClient(certFolder, hostPort)
}

// StartHubMessenger creates and starts a messenger from the hub configuration
// If the messenger can't connect it will try forever or until Disconnect() is called.
//  clientID is used to identify the client to the message bus and must be unique
//  config contains the connection information of the hub
// returns a messenger used to publish and subscribe to channels.
func StartHubMessenger(clientID string, hubConfig *config.HubConfig) (IHubMessenger, error) {
	messenger := NewHubConnection(
		hubConfig.Messenger.Protocol,
		hubConfig.Messenger.CertFolder, hubConfig.Messenger.HostPort,
	)
	err := messenger.Connect(clientID, hubConfig.Messenger.Timeout)
	return messenger, err
}
