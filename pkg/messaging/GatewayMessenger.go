package messaging

import (
	"github.com/wostzone/gateway/pkg/messaging/mqtt"
	"github.com/wostzone/gateway/pkg/messaging/smbus"
)

// ConnectionProtocol of the gateway
type ConnectionProtocol string

// Available message bus connection protocols
const (
	ConnectionProtocolSmbus ConnectionProtocol = "smb"
	ConnectionProtocolMQTT  ConnectionProtocol = "mqtt"
)

// NewGatewayConnection creates a messenger to a WoST gateway message bus
// protocol selects the connection method used by the gateway: "" for default, "mqtt"
// hostPort is the hostname and port of the messaging server to connect to
// certFolder is the folder containing the client and server certificates for TLS connections
//       Leave empty to connect to a gateway that is not using TLS, only for testing
//       Both client and server certificates are used for two-way authentication
// Returns nil if the connection method is unknown
func NewGatewayConnection(protocol ConnectionProtocol, certFolder string, hostPort string) IGatewayMessenger {
	if protocol == ConnectionProtocolSmbus {
		return smbus.NewSmbusMessenger(certFolder, hostPort)
	}
	if protocol == ConnectionProtocolMQTT {
		return mqtt.NewMqttMessenger(certFolder, hostPort)
	}
	return nil
}
