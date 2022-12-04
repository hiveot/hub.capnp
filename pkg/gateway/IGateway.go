package gateway

import (
	"context"

	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub/internal/caphelp"
)

const ServiceName = "gateway"

// ConfigCapabilitiesSection is the name of the capabilities section in the config file
const ConfigCapabilitiesSection = "capabilities"

// ConfigClientTypeField name of the field that holds the type of clients allowed to use the capability
const ConfigClientTypeField = "clientType"

// GatewayInfo describes the gateway's capabilities and capacity
//type GatewayInfo struct {
//	// Capabilities describes the capabilities available via this gateway
//	Capabilities []caphelp.CapabilityInfo
//
//	// URL to the URL to the gateway service that provides this capability.
//	// The gateway implements the IGateway interface that can be used to obtain access to the capability.
//	// A local UDS socket: uds://run/name.socket
//	// A remote Websocket service: wss://address:port/name
//	URL string `json:"url"`
//
//	// Latency of reaching the capability in msec
//	Latency int `json:"latency"`
//
//	// current CPU utilization of the host
//	//CPUUtilization int `json:"CPUUtilization"`
//
//	// memory utilization of the host
//
//	// cpu capacity of the host
//
//	// memory capacity of the host
//}

// IGatewayService provides Hub capabilities that are available on the device to clients such as IoT devices, services and
// end-users.
type IGatewayService interface {
	// GetCapability returns the capnp capability by name
	// The client login determines what capabilities are available.
	// TODO: automatically determine clientID and type when connected through UDS or TLS with client cert
	// in the meantime, pass it in manually
	//   clientID of the connected client
	//   clientType of the connected client
	GetCapability(ctx context.Context, clientID, clientType, capabilityName string, args []string) (
		capability capnp.Client, err error)

	// ListCapabilities returns the aggregated list of capabilities from all connected services
	// This list is reduced to capabilities based on the client type.
	ListCapabilities(_ context.Context, clientType string) (infoList []caphelp.CapabilityInfo, err error)

	// Login to the gateway in order to get additional capabilities
	// Login detects the client type (service, iotdevice, user) based on the connection method.
	Login(ctx context.Context, clientID string, password string) (success bool, err error)

	// Ping helps determine if the service is reachable
	Ping(ctx context.Context) (reply string, err error)
}
