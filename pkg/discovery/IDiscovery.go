package discovery

import "github.com/hiveot/hub/pkg/gateway"

// IDiscovery defines the interface to the discovery service.
// This extends the Gateway to include remote capabilities.
type IDiscovery interface {

	// GetCapability provides the capability with the highest QoS
	// This first determines the device to use, using the service.
	// Then in the client obtains and returns the remote capability.
	// 'in the client' means that the service does not proxy the capability, avoiding an unnecessary hop.
	//
	//  clientType is the type of authenticated client
	//  name is the name of the capability as provided in capability info. Usually the service name[/nestedCap].
	GetCapability(clientType string, name string) interface{}

	// AvailableCapabilities lists the capabilities available to the client
	// This list varies depending on whether the client is authenticated, is an IoT device, service or end-user.
	// The discovery service can include local and remote capabilities.
	//  clientType to return the capabilities for the type of client. See constants above.
	AvailableCapabilities(localOnly bool, clientType string) []*gateway.GatewayInfo
}
