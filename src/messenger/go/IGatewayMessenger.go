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

// IGatewayMessenger interface to connection handler to publish messages onto and subscribe to the gateway
type IGatewayMessenger interface {

	// Connect the messenger to the messenger server.
	// If a connection is already in place, the existing connection is dropped and
	// a new connection is made. Existing subscriptions remain in place.
	//  hostPort contains the hostname or ip address of the message bus with the port
	//  clientID is unique to the server
	//  timeout is the amount of time to keep retrying in case the connection fails
	Connect(hostPort string, clientID string, timeout int) error

	// Disconnect all connections and remove all subscriptions
	Disconnect()

	// Publish sends a message to the gateway on the given topic
	// topic contains the address to publish to, divided by '/' as a separator
	Publish(topic string, message []byte) error

	// Subscribe to a message topic.
	// Only a single subscription to a topic can be made.
	// topic contains the address to listen on. Wildcard supports depends on the messenger used.
	// handler is invoked when a message is received on the address
	Subscribe(topic string, handler func(address string, message []byte))

	// Unsubscribe from a previously subscribed topic address.
	// topic contains the address to listen on, divided by '/' as a separator
	Unsubscribe(topic string)
}
