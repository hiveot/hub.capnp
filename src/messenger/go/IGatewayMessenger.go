// Package messenger with interface for gateway connection management
// Used by implementation of the service bus connection to communicate between plugin and gateway
package messenger

// Default client and server certificates names when used
const (
	CaCertFile     = "ca.crt"
	CaKeyFile      = "ca.key"
	ServerCertFile = "gateway.crt"
	ServerKeyFile  = "gateway.key"
	ClientCertFile = "client.crt"
	ClientKeyFile  = "client.key"
)

// Predefined gateway channels
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

// IGatewayMessenger interface to connection handler to publish messages onto and subscribe to the gateway
type IGatewayMessenger interface {

	// Connect the messenger to the messenger server.
	// If a connection is already in place, the existing connection is dropped and
	// a new connection is made. Existing subscriptions remain in place.
	//  hostPort contains the hostname or ip address of the message bus with the port
	//  clientID is unique to the server. Default is hostname-timestamp
	//  timeout is the amount of time to keep retrying in case the connection fails
	Connect(hostPort string, clientID string, timeout int) error

	// Disconnect all connections and remove all subscriptions
	Disconnect()

	// Publish sends a message to the gateway on the given channelID
	// channelID contains the address to publish to, divided by '/' as a separator
	Publish(channelID string, message []byte) error

	// Subscribe to a message channelID.
	// Only a single subscription to a channelID can be made.
	// channelID contains the address to listen on. Wildcard supports depends on the messenger used.
	// handler is invoked when a message is received on the address
	Subscribe(channelID string, handler func(address string, message []byte))

	// Unsubscribe from a previously subscribed channelID address.
	// channelID contains the address to listen on, divided by '/' as a separator
	Unsubscribe(channelID string)
}
