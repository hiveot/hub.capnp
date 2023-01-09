package resolver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

const ServiceName = hubapi.ResolverServiceName

const DefaultResolverPath = hubapi.DefaultResolverAddress

// IResolverService lists all available capabilities
type IResolverService interface {

	// ListCapabilities returns the list of capabilities provided by capability providers.
	ListCapabilities(ctx context.Context, clientType string) (capInfo []CapabilityInfo, err error)

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
	//RegisterCapabilities(ctx context.Context, providerID string, capInfo []CapabilityInfo, provider hubapi.CapProvider) error

}
