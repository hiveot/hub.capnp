// Package plugin with interface for gateway connection management
// Used by implementation of the service bus connection to communicate between plugin and gateway
package plugin

// Default client and server certificates names when used
const (
	CaCertFile     = "ca.crt"
	CaKeyFile      = "ca.key"
	ServerCertFile = "gateway.crt"
	ServerKeyFile  = "gateway.key"
	ClientCertFile = "client.crt"
	ClientKeyFile  = "client.key"
)

// IGatewayMessenger interface to connection handler to publish messages onto and subscribe to the gateway
type IGatewayMessenger interface {

	// Connect the messenger to the messenger server
	Connect(serverAddress string) error

	// Disconnect all connections and stop listeners
	Disconnect()

	// Publish sends a message to the gateway on the given topic
	// This automatically creates a new connection to the server bus to publish onto.
	// topic contains the address to publish to, divided by '/' as a separator
	Publish(topic string, message []byte) error

	// Subscribe to a message topic. This creates a new connection to the service bus
	// to listen on.
	// topic contains the address to listen on, divided by '/' as a separator
	// handler is invoked when a message is received on the address
	Subscribe(topic string, handler func(address string, message []byte)) error
}
