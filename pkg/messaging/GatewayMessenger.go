package messaging

import (
	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/pkg/messaging/mqtt"
	"github.com/wostzone/gateway/pkg/messaging/smbclient"
)

// ConnectionProtocol of the gateway
// type ConnectionProtocol string

// Available message bus connection protocols
const (
	ConnectionProtocolSmbus string = "smbus"
	ConnectionProtocolMQTT  string = "mqtt"
)

// NewGatewayConnection creates a messenger to a WoST gateway message bus.
//
// TLS is disabled if certFolder is empty. smbus generates both server and client certificates if
// the folder is empty. They are used for two-way authentication.
// If the given protocol is unknown, smbus will be used.
//
//  protocol selects the connection method used by the gateway: "mqtt" or "smbus".
//  certFolder is the folder containing the client and server certificates for TLS connections
//  hostPort is the hostname and port of the messaging server to connect to
// Returns nil if the connection method is unknown
func NewGatewayConnection(protocol string, certFolder string, hostPort string) IGatewayMessenger {
	if protocol == ConnectionProtocolMQTT {
		return mqtt.NewMqttMessenger(certFolder, hostPort)
	}
	return smbclient.NewSmbClient(certFolder, hostPort)
}

// StartGatewayMessenger creates and starts a messenger from the gateway configuration
// If the messenger can't connect it will try forever or until Disconnect() is called.
//  clientID is used to identify the client to the message bus and must be unique
//  config contains the connection information of the gateway
// returns a messenger used to publish and subscribe to channels.
func StartGatewayMessenger(clientID string, gwConfig *lib.GatewayConfig) (IGatewayMessenger, error) {
	messenger := NewGatewayConnection(
		gwConfig.Messenger.Protocol,
		gwConfig.Messenger.CertFolder, gwConfig.Messenger.HostPort,
	)
	err := messenger.Connect(clientID, gwConfig.Messenger.Timeout)
	return messenger, err
}
