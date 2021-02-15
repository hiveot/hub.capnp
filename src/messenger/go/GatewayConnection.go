package messenger

// ConnectionProtocol of the gateway
type ConnectionProtocol string

// Available message bus connection protocols
const (
	ConnectionProtocolISB  ConnectionProtocol = "isb"
	ConnectionProtocolMQTT ConnectionProtocol = "mqtt"
)

// NewGatewayConnection creates a messenger to a WoST gateway message bus
// protocol selects the connection method used by the gateway: "" for default, "mqtt"
// certFolder is the folder containing the client and server certificates for TLS connections
//       Leave empty to connect to a gateway that is not using TLS, only for testing
//       Both client and server certificates are used for two-way authentication
// Returns nil if the connection method is unknown
func NewGatewayConnection(protocol ConnectionProtocol, certFolder string) IGatewayMessenger {
	if protocol == ConnectionProtocolISB {
		return NewISBMessenger(certFolder)
	}
	if protocol == ConnectionProtocolMQTT {
		return NewMqttMessenger(certFolder)
	}
	return nil
}
