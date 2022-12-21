package resolver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

const ServiceName = "resolver"
const DefaultResolverPath = hubapi.DefaultResolverAddress

// CapabilityInfo provides information on a capability that is available through the resolver and gateways
type CapabilityInfo struct {

	// Name of the capability. This is the capnp interface name as defined by the service.
	CapabilityName string

	// list of arguments that are required. TBD.
	CapabilityArgs []string

	// Type of clients that can use the capability. See ClientTypeXyz above
	ClientTypes []string

	// ServiceID of the service providing the capability.
	ServiceID string

	// Network is provided when the capability is available via a direct connections, bypassing the resolver.
	// "unix" for Unix Domain sockets and 'tcp' for TCP sockets. Default is unix.
	DNetwork string

	// Address is provided when the capability is available via a direct connections, bypassing the resolver.
	// This endpoint at this address implements the IResolveCapability interface.
	// The address format depends on the network and protocol.
	// * leave empty to not allow direct connections (default)
	// * unix networks provide the socket path to connect to the service providing the capability.
	// * tcp networks provide the IP address:port, and optionally a path, depending on the protocol
	DAddress string

	// Protocol indicates what protocol to use to get the capability
	// The default is 'capnp' protocol. Other services can use protocols such as https and rtsp.
	Protocol string
}

// IResolverService is a resolver for capabilities.
type IResolverService interface {

	// OnIncomingConnection notifies the service of a new incoming connection.
	// This is invoked by the underlying protocol and returns a new session to use
	// with the connection.
	// If this connection closes then capabilites added in this session are removed.
	OnIncomingConnection(conn net.Conn) IResolverSession

	// OnConnectionClosed is invoked if the connection with the client has closed.
	// The service will remove the session.
	OnConnectionClosed(conn net.Conn, session IResolverSession)

	// Stop the service and disconnect from the resolver
	//Stop()
}

// IResolverSession is a client of the resolver service using to access and register capabilities
type IResolverSession interface {

	// GetCapability is used to obtain the capability using the capnp protocol.
	// If Level 3 RPC is support, this might hand the capability over to the client.
	// The provided capability must be released after use.
	GetCapability(ctx context.Context,
		clientID string, clientType string, capName string, args []string) (capnp.Client, error)

	// ListCapabilities returns the list of capabilities provided by capability providers.
	ListCapabilities(ctx context.Context) (capInfo []CapabilityInfo, err error)

	// RegisterCapabilities is invoked by capability providers to register their capabilities
	// along with a callback to retrieve a capability. Capability provides can also provide
	// capabilities from other services and act as a proxy.
	//
	// The service provider associates the capabilities with the connection and will remove them
	// when the connection is broken.
	//
	// This must be invoked each time the available capabilities change. For example,the gateway
	// service provides remote capabilities that can changes when a connection to a remote service
	// is made or broken.
	// This returns if the provider or capInfo are invalid
	RegisterCapabilities(ctx context.Context, providerID string, capInfo []CapabilityInfo, provider hubapi.CapProvider) error

	// Release is defined in ICapProvider
}

// IProvider provides capabilities from service providers.
// This is the callback in to RegisterCapabilities to provide the service capabilities
//type IProvider interface {
//
//	// GetCapability is used to obtain the capability using the capnp protocol.
//	// If Level 3 RPC is support, this might hand the capability over to the client.
//	// The provided capability must be released after use.
//	GetCapability(ctx context.Context,
//		clientID string, clientType string, capName string, args []string) (capnp.Client, error)
//
//	// ListCapabilities returns the list of capabilities provided by capability providers.
//	ListCapabilities(ctx context.Context) (capInfo []CapabilityInfo, err error)
//
//	// Release the connection to the provider
//	Release()
//}
