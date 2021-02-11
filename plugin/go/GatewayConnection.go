package plugin

// Predefined channels
const (
	// The TD channel carries 'Thing Description' documents
	TDChannelID = "TD"
	// The notification channel carries Thing status updates
	NotificationChannelID = "notification"
	// The action channel carries Thing action commands
	ActionChannelID = "action"
	// The plugin channel carries plugin registration messages
	PluginChannelID = "plugin"
	// The test channel carries test messages
	TestChannelID = "test"
)

// ConnectionProtocol of the gateway
type ConnectionProtocol string

// Available message bus connection protocols
const (
	ConnectionProtocolISB  ConnectionProtocol = "isb"
	ConnectionProtocolMQTT ConnectionProtocol = "mqtt"
)

// NewGatewayConnection creates a messenger to a WoST gateway message bus
// pluginID is the unique ID of the plugin. If multiple instances of a plugin are used
//          then each must have a unique ID. It is recommended to use the plugin type
//          name followed by an instance name. For example openzwave-1, openzwave-2.
// protocol selects the connection method used by the gateway: "" for default, "mqtt"
// certFolder is the folder containing the client and server certificates for TLS connections
//       Leave empty to connect to a gateway that is not using TLS, only for testing
//       Both client and server certificates are used for two-way authentication
// Returns nil if the connection method is unknown
func NewGatewayConnection(pluginID string, protocol ConnectionProtocol, certFolder string) IGatewayMessenger {
	if pluginID == "" {
		return nil
	}
	if protocol == ConnectionProtocolISB {
		return NewISBMessenger(pluginID, certFolder)
	}
	if protocol == ConnectionProtocolMQTT {
		return NewMqttMessenger(pluginID, certFolder)
	}
	return nil
}
