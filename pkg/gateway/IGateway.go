package gateway

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/resolver"
)

const ServiceName = hubapi.GatewayServiceName

// ClientInfo contains client info as seen by the gateway
// Intended for diagnostics and troubleshooting
type ClientInfo struct {
	// ClientID that is connected. loginID, serviceID, or IoT device ID
	ClientID string

	// ClientType identifies how the client is authenticated. See also the resolver:
	//  ClientTypeUnauthenticated   - client is not authenticated
	//  ClientTypeUser              - client is authenticated as a user with login/password
	//  ClientTypeIoTDevice         - client is authenticated as an IoT device with certificate
	//  ClientTypeService           - client is authenticated as a service with certificate
	ClientType string
}

// IGatewayService provides the capability to accept new sessions with remote clients
//type IGatewayService interface {
//	// OnIncomingConnection notifies the service of a new incoming RPC connection.
//	// This is invoked by the underlying RPC protocol (eg capnp) server.
//	// This creates a new session for each connection in order to track authentication
//	// and performance. If the RPC connection closes the session is released.
//	OnIncomingConnection(conn net.Conn) IGatewaySession
//
//	// OnConnectionClosed is invoked if the connection with the client has closed.
//	// The service will remove the session.
//	OnConnectionClosed(conn net.Conn, session IGatewaySession)
//}

// IGatewaySession provides Hub capabilities to clients on the network
// Each client connection receives a session with the capabilities that are dependent
// on the client's authentication.
type IGatewaySession interface {

	// ListCapabilities returns the list of capabilities provided by capability providers.
	ListCapabilities(ctx context.Context) (capInfo []resolver.CapabilityInfo, err error)

	// Login to the gateway as a user in order to get additional capabilities.
	// This returns an authToken and refreshToken that can be used with services that require
	// authentication.
	// If the authentication token has expired then call refresh.
	Login(ctx context.Context, clientID, password string) (authToken, refreshToken string, err error)

	// Ping helps determine if the gateway is reachable
	Ping(ctx context.Context) (reply ClientInfo, err error)

	// Refresh the token pair
	Refresh(ctx context.Context, oldRefreshToken string) (newAuthToken, newRefreshToken string, err error)

	// Release the session when its incoming RPC connection closes
	Release()
}
