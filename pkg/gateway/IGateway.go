package gateway

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/resolver"
)

const ServiceName = hubapi.GatewayServiceName

// HIVEOT_DNSSD_TYPE service type for hiveot services
const HIVEOT_DNSSD_TYPE = "_hiveot._tcp"

// ClientInfo contains client info as seen by the gateway
// Intended for diagnostics and troubleshooting
type ClientInfo struct {
	// ClientID that is connected. loginID, serviceID, or IoT device ID
	ClientID string

	// AuthType identifies how the client is authenticated. See also the resolver:
	//  AuthTypeUnauthenticated   - client is not authenticated
	//  AuthTypeUser              - client is authenticated as a user with login/password
	//  AuthTypeIoTDevice         - client is authenticated as an IoT device with certificate
	//  AuthTypeService           - client is authenticated as a service with certificate
	AuthType string
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
	//
	// If successful this sets the session clientID to the given client ID and
	// sets the session to authenticated.
	//
	// This returns an authToken and refreshToken that can be used with services that require
	// authentication. The refresh token is valid for N days where N is configured
	// in the service. Default is defined in authn and is 14 days.
	// The refresh token can be used with 'Refresh' to reauthenticate in a new sessions
	// as long as the token is still valid.
	Login(ctx context.Context, clientID, password string) (authToken, refreshToken string, err error)

	// Ping helps determine if the gateway is reachable
	Ping(ctx context.Context) (reply ClientInfo, err error)

	// Refresh the auth token pair and reauthenticates the session.
	// The token must be for the given clientID and must still be valid.
	// This returns a new refresh token that is valid for another N days, where N is configured
	// in the service.
	Refresh(ctx context.Context, clientID string, oldRefreshToken string) (newAuthToken, newRefreshToken string, err error)

	// Release the session when its incoming RPC connection closes
	Release()
}
