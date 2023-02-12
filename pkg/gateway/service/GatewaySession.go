package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
	capnpclient2 "github.com/hiveot/hub/pkg/authn/capnpclient"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
	"github.com/hiveot/hub/pkg/resolver/service"
)

// GatewaySession implements the IGatewaySession interface.
// A new instance is created by the capnp server for each incoming connection.
// This session is intended as a proxy for remote services to the local resolver.
type GatewaySession struct {

	// ID of the authenticated client. Set when authentication is successful.
	clientID string

	// client connection
	clientConn *tls.Conn

	// type of the connected client
	authType string

	//
	resolverPath    string
	resolverService *capnpclient.ResolverServiceCapnpClient
	//resolverClient  *hubapi.CapResolverService
	resolverConn net.Conn

	// Cached user authn capability for login and refresh
	authnService authn.IAuthnService
	userAuthn    authn.IUserAuthn
}

// Provide the user authentication service
func (session *GatewaySession) getUserAuthn(
	ctx context.Context, clientID string) (userAuthn authn.IUserAuthn, err error) {

	if session.userAuthn == nil {
		// if the authn service is available ask for the capability, otherwise ask the resolver
		// intended for testing
		if session.authnService != nil {
			// the resolver capnp client is a proxy for all capabilities it has a connection to
			session.userAuthn, err = session.authnService.CapUserAuthn(ctx, clientID)
		} else {
			// the resolver capnp client is a proxy for all capabilities it has a connection to
			capAuthn := hubapi.CapAuthn(session.resolverService.Capability())
			authnClient := capnpclient2.NewAuthnClientFromCapnpCapability(capAuthn)
			session.userAuthn, err = authnClient.CapUserAuthn(ctx, clientID)
		}
		if err != nil {
			err = fmt.Errorf("can't connect to the authn service: %s", err)
			logrus.Error(err)
		}
	}
	return session.userAuthn, err

}

// HandleUnknownMethod forwards the request to the resolver.
func (session *GatewaySession) HandleUnknownMethod(m capnp.Method) *server.Method {
	reject := true
	// Check the available capabilities for this client
	capList, err := session.resolverService.ListCapabilities(context.Background(), session.authType)
	if err == nil {
		for _, capInfo := range capList {
			if capInfo.InterfaceID == m.InterfaceID && capInfo.MethodID == m.MethodID {
				reject = false
				break
			}
		}
	}
	if reject {
		logrus.Warningf("client '%s' of type '%s' is not allowed to invoke of InterfaceID=%x, MethodID=%x",
			session.clientID, session.authType, m.InterfaceID, m.MethodID)
		return nil
	}
	// return a helper for forwarding the request to the resolver
	capResolverClient := capnp.Client(session.resolverService.Capability())
	forwarderMethod := service.NewForwarderMethod(m, &capResolverClient)
	return forwarderMethod
}

// ListCapabilities returns list of capabilities of all connected services sorted by service and capability names
func (session *GatewaySession) ListCapabilities(ctx context.Context) ([]resolver.CapabilityInfo, error) {
	capList := make([]resolver.CapabilityInfo, 0)
	//cstate := session.clientConn.ConnectionState()
	//logrus.Infof("clientID: %v", cstate.PeerCertificates[0].Subject.CommonName)
	//logrus.Infof("authTypes: %v", cstate.PeerCertificates[0].Subject.OrganizationalUnit)
	//logrus.Infof("handshake: %v", cstate.HandshakeComplete)
	//logrus.Infof("clientID='%s'", session.clientID)
	capList, err := session.resolverService.ListCapabilities(ctx, session.authType)
	return capList, err
}

// Login to the gateway
// if no userauthn service is available then refuse
// This sets the session clientID to the given ID when successful
func (session *GatewaySession) Login(ctx context.Context, clientID, password string) (
	authToken string, refreshToken string, err error) {

	// need authn capability to login
	userAuthn, err := session.getUserAuthn(ctx, clientID)
	if err != nil {
		return "", "", err
	}
	authToken, refreshToken, err = userAuthn.Login(ctx, password)

	if err == nil {
		logrus.Infof("Login of user '%s' successful.", clientID)
		session.authType = hubapi.AuthTypeUser
		session.clientID = clientID
	} else {
		session.authType = hubapi.AuthTypeUnauthenticated
		err = fmt.Errorf("login of '%s' failed: %s", clientID, err)
		logrus.Warning(err)
	}
	return authToken, refreshToken, err
}

// Ping capability
func (session *GatewaySession) Ping(_ context.Context) (gateway.ClientInfo, error) {
	logrus.Infof("Ping")
	ci := gateway.ClientInfo{
		ClientID: session.clientID,
		AuthType: session.authType,
	}
	return ci, nil
}

// Refresh authentication tokens
func (session *GatewaySession) Refresh(ctx context.Context, clientID string, oldRefreshToken string) (
	authToken string, refreshToken string, err error) {

	// use authn capability to refresh
	userAuthn, err := session.getUserAuthn(ctx, clientID)
	if err != nil {
		return "", "", err
	}
	authToken, refreshToken, err = userAuthn.Refresh(ctx, oldRefreshToken)
	return authToken, refreshToken, err
}

// Release the connection to the resolver
func (session *GatewaySession) Release() {
	logrus.Infof("releasing session of client '%s'", session.clientID)
	if session.resolverService != nil {
		session.resolverService.Release()
	}
	if session.resolverConn != nil {
		_ = session.resolverConn.Close() // is this needed?
	}
	// do not release the userAuthn service as it belongs to the service, not the session
}

// StartGatewaySession creates a new gateway session with the resolver to serve gateway requests.
// Use Release after the remote connection to the gateway is closed.
// This returns an error if connecting with the resolver fails.
// The user authentication is on loan to the session and should not be released.
//
//	resolverPath is the socket address for the resolver
//	clientID is the client ID
//	      use "" if the client is not authenticated and must use Login or Refresh
//	authType is the authentication type of the client, e.g. unauth, service, device or user.
//	      using hubapi.AuthTypeUnauthenticated when not authenticated
//	clientConn is the TLS connection with the client. This will be closed on release
//	authnService optional authentication service. Intended for testing.
func StartGatewaySession(
	resolverPath string, clientID string, authType string, clientConn *tls.Conn,
	authnService authn.IAuthnService) (*GatewaySession, error) {

	ctx := context.Background()
	session := &GatewaySession{
		clientID:     clientID,
		clientConn:   clientConn,
		authType:     authType,
		resolverPath: resolverPath,
		authnService: authnService,
	}
	if clientConn != nil {

	}
	resolverConn, err := net.Dial("unix", resolverPath)
	if err != nil {
		err = fmt.Errorf("unable to connect to the resolver socket at '%s': %s", resolverPath, err)
		return nil, err
	}
	session.resolverConn = resolverConn
	session.resolverService, err = capnpclient.NewResolverServiceCapnpClient(ctx, resolverConn)

	if err != nil {
		_ = resolverConn.Close()
		return nil, err
	}

	return session, err
}
