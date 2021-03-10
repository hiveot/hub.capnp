// Package messaging with interface for hub connection management
// Used by implementation of the service bus connection to communicate between plugin and hub
package messaging

// Default client and server certificates names when used
const (
	CaCertFile     = "ca.crt"
	CaKeyFile      = "ca.key"
	ServerCertFile = "hub.crt"
	ServerKeyFile  = "hub.key"
	ClientCertFile = "client.crt"
	ClientKeyFile  = "client.key"
)

// Predefined hub channels
const (
	// The TD channel carries 'Thing Description' documents
	TDChannelID = "td"
	// The notification channel carries Thing status updates
	EventsChannelID = "events"
	// The action channel carries Thing action commands
	ActionChannelID = "actions"
	// The plugin channel carries plugin registration messages
	PluginsChannelID = "plugin"
	// The test channel carries test messages
	TestChannelID = "test"
)

// PluginMessage with activate/deactivation message for the PluginChannel
type PluginMessage struct {
	Active   bool   `json:"active"`   // Plugin is active/inactive
	ID       string `json:"ID"`       // Plugin instance ID that started
	Hostname string `json:"hostname"` // Hostname where it can be reached at if applicable
	IP4      string `json:"ip4"`      // IP4 address where it can be reached at if applicable
	Port     int    `json:"port"`     // Port the service is listening on
}

// IHubMessenger interface to connection handler to publish messages onto and subscribe to the hub
type IHubMessenger interface {

	// Connect the messenger to the messenger server.
	// If a connection is already in place, the existing connection is dropped and
	// a new connection is made. Existing subscriptions remain in place.
	// The messenger must already be setup with the destination host, port and certificate
	//  clientID is unique to the server. Default is hostname-timestamp
	//  timeout is the amount of time to keep retrying in case the connection fails. 0 to try indefinitly
	Connect(clientID string, timeout int) error

	// Disconnect all connections and remove all subscriptions
	Disconnect()

	// Publish sends a message to the hub on the given channelID
	//  channelID contains the address to publish to, divided by '/' as a separator
	Publish(channelID string, message []byte) error

	// Subscribe to a message channelID.
	// Only a single subscription to a channelID can be made.
	//  channelID contains the address to listen on. Wildcard supports depends on the messenger used.
	//  handler is invoked when a message is received on the channel
	Subscribe(channelID string, handler func(channelID string, message []byte))

	// Unsubscribe from a previously subscribed channelID address.
	//  channelID contains the address to listen on, divided by '/' as a separator
	Unsubscribe(channelID string)
}
