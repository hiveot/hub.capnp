package gateway

import (
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
type IGatewayService interface {
	//	// OnIncomingConnection notifies the service of a new incoming RPC connection.
	//	// This is invoked by the underlying RPC protocol (eg capnp) server.
	//	// This creates a new session for each connection in order to track authentication
	//	// and performance. If the RPC connection closes the session is released.
	//	OnIncomingConnection(conn net.Conn) IGatewaySession
	//
	//	// OnConnectionClosed is invoked if the connection with the client has closed.
	//	// The service will remove the session.
	//	OnConnectionClosed(conn net.Conn, session IGatewaySession)

	// NewSession returns a hub gateway session using an authentication token
	//  sessionToken can be obtained using any of the service auth methods.
	// This returns a capability provider for capabilities that are available
	// to the client based on their authentication method and clientID.
	// The CapProvider client
	NewSession(sessionToken string) (resolver.ICapProvider, error)

	// AuthNoAuth returns a session token for unauthenticated users
	AuthNoAuth(clientID string) (sessionToken string)

	// AuthProxy returns a session token for clients of a proxy services
	// the proxy service MUST be authenticated using a client certificate or
	// no token will be returned.
	AuthProxy(clientID string, clientCertPEM string) (sessionToken string)

	// AuthRefresh issues a new session token using an existing non-expired token
	// Intended for resuming a session without requiring a new login.
	// The old session token can be invalidated and should no longer be used.
	AuthRefresh(clientID string, oldSessionToken string) (sessionToken string)

	// AuthWithCert obtains a session token for clients that connect with a client certificate.
	AuthWithCert() (sessionToken string)

	// AuthWithPassword obtains a session token for users with login id and password
	AuthWithPassword(loginID string, password string) (sessionToken string)
}
