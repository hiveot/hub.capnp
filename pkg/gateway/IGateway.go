package gateway

import "context"

const ServiceName = "gateway"

type IService interface{}

// ConfigCapabilitiesSection is the name of the capabilities section in the config file
const ConfigCapabilitiesSection = "capabilities"

// ConfigClientTypeField name of the field that holds the type of clients allowed to use the capability
const ConfigClientTypeField = "clientType"

// type of capability
const (
	// ClientTypeUnauthenticated for clients without authentication
	ClientTypeUnauthenticated = "noauth"

	// ClientTypeIotDevice for clients authenticated as IoT devices
	ClientTypeIotDevice = "iotdevice"

	// ClientTypeUser for clients authenticated as end-users
	ClientTypeUser = "user"

	// ClientTypeService for clients authenticated as Hub services
	ClientTypeService = "service"
)

// CapabilityInfo provides information on a capabilities available through the gateway
type CapabilityInfo struct {

	// Service name that is providing the capability.
	Service string `json:"service"`

	// Name of the capability. This is the method name as defined by the service.
	Name string `json:"name"`

	// Type of clients that can use the capability. See ClientTypeXyz above
	ClientType []string `json:"clients"`
}

// GatewayInfo describes the gateway's capabilities and capacity
type GatewayInfo struct {
	// Capabilities describes the capabilities available via this gateway
	Capabilities []CapabilityInfo

	// URL to the URL to the gateway service that provides this capability.
	// The gateway implements the IGateway interface that can be used to obtain access to the capability.
	// A local UDS socket: uds://run/name.socket
	// A remote Websocket service: wss://address:port/name
	URL string `json:"url"`

	// Latency of reaching the capability in msec
	Latency int `json:"latency"`

	// current CPU utilization of the host
	//CPUUtilization int `json:"CPUUtilization"`

	// memory utilization of the host

	// cpu capacity of the host

	// memory capacity of the host
}

// IGatewayService provides Hub capabilities that are available on the device to clients such as IoT devices, services and
// end-users.
type IGatewayService interface {

	// GetCapability obtains the capability with the given name, if available
	//
	// This returns the client for that capability, or nil if the capability is not available. The result must be
	// cast to the corresponding interface.
	//
	// All capabilities must be released after use. If the gateway capability is released or disconnected, then
	// all capabilities obtained via the gateway are also released.
	//
	// The capabilities that are available depend on how the client is authenticated at and whether it is
	// a device, service or end-user client.
	//
	//  clientType is the type of authenticated client
	//  service is the name of the service providing capabilities
	GetCapability(ctx context.Context, clientType string, service string) (interface{}, error)

	// GetGatewayInfo describes the capabilities and capacity of the gateway
	GetGatewayInfo(ctx context.Context) (GatewayInfo, error)

	// Ping helps determine if the service is reachable
	Ping(ctx context.Context) (string, error)

	// Stop the service and free its resources
	Stop(ctx context.Context) error
}
