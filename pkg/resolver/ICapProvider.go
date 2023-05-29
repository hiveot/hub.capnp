package resolver

import (
	"context"
)

// CapabilityInfo provides information on obtaining a capability from a service.
// This does not describe the capability itself. It is up to the caller to apply the
// correct interface to the provided capability.
type CapabilityInfo struct {
	// Internal capnp ID of the interface that provides the capability.
	// This is typically the bootstrap interface of the service providing the method to get the capability.
	InterfaceID uint64

	// ID of the interfaceInternal capnp ID of the interface that provides the capability.
	// This is typically the bootstrap interface of the service providing the method to get the capability.
	InterfaceName string

	// Internal capnp method ID of the method that provides the capability.
	// This is the method index in the bootstrap interface above.
	MethodID uint16

	// MethodName is the canonical name of the method in the bootstrap interface that
	// provides the capability. Method names must be unique.
	MethodName string

	// Type of authentication that is allowed to use the capability. See hubapi.AuthTypeXyz
	AuthTypes []string

	// Protocol indicates what protocol to use to get the capability
	// The default is 'capnp' protocol. Other services can use protocols such as https and rtsp.
	Protocol string

	// ServiceID of the service providing the capability. Used to connect to the service.
	ServiceID string

	// Network is the direct connection network to use for connection with the service.
	// "unix" for Unix Domain sockets and 'tcp' for TCP sockets. Default is unix.
	Network string

	// Address is the connection address of the service implementing the interface and method
	// to obtain the capability.
	// * leave empty to use the connection that provided this info (default)
	// * unix domain sockets provide the socket path to dial into.
	// * tcp networks provide the IP address:port, and optionally a path, depending on the protocol
	Address string
}

// ICapProvider is the native interface of a capability provider used to provide capabilities
// This is typically not used directly but through the local resolver client.
type ICapProvider interface {

	// ListCapabilities returns the list of capabilities provided by capability providers.
	ListCapabilities(ctx context.Context) (capInfo []CapabilityInfo, err error)

	// Release must be called after to close the session
	Release()
}
