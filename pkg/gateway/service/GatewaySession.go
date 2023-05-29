package service

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	"context"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
	"github.com/hiveot/hub/pkg/resolver/service"
)

// GatewaySession implements the resolver.ICapProvider interface.
//
// The purpose of the gateway session is to proxy capability requests to the providing services.
// This utilizes the resolver service which knows what service provides what capability.
//
// The available capabilities are obtained from the resolver service and limited to
// those matching the authType (device, user, service).
type GatewaySession struct {

	// ID of the authenticated client. Set when authentication is successful.
	clientID string

	// the capabilities available to this session
	capList []resolver.CapabilityInfo

	// client connection
	//clientConn *tls.Conn

	// Type of the connected client
	authType string

	//
	//resolverPath    string
	//resolverService *capnpclient.ResolverCapnpClient
	//resolverClient  capnp.Client
	//capResolverService  hubapi.CapResolverService
	capResolverService capnp.Client

	// Cached user authn capability for login and refresh
	//authnService authn.IAuthnService
	//userAuthn    authn.IUserAuthn
}

// Provide the user authentication service
//func (session *GatewaySession) getUserAuthn(
//	ctx context.Context, clientID string) (userAuthn authn.IUserAuthn, err error) {
//
//	if session.userAuthn == nil {
//		// if the authn service is available ask for the capability, otherwise ask the resolver
//		// intended for testing
//		if session.authnService != nil {
//			// the resolver capnp client is a proxy for all capabilities it has a connection to
//			session.userAuthn, err = session.authnService.CapUserAuthn(ctx, clientID)
//		} else {
//			// the resolver capnp client is a proxy for all capabilities it has a connection to
//			capAuthn := capnp.Client(session.resolverService.Capability())
//			authnClient := capnpclient2.NewAuthnCapnpClient(capAuthn)
//			session.userAuthn, err = authnClient.CapUserAuthn(ctx, clientID)
//		}
//		if err != nil {
//			err = fmt.Errorf("can't connect to the authn service: %s", err)
//			logrus.Error(err)
//		}
//	}
//	return session.userAuthn, err
//
//}

// HandleUnknownMethod forwards the request to the resolver.
func (session *GatewaySession) HandleUnknownMethod(m capnp.Method) *server.Method {
	reject := true
	// Check the available capabilities for this client
	for _, capInfo := range session.capList {
		if capInfo.InterfaceID == m.InterfaceID && capInfo.MethodID == m.MethodID {
			reject = false
			break
		}
	}

	if reject {
		logrus.Warningf("client '%s' of type '%s' is not allowed to invoke of InterfaceID=%x, MethodID=%x",
			session.clientID, session.authType, m.InterfaceID, m.MethodID)
		return nil
	}
	// return a helper for forwarding the request to the resolver
	//capResolverClient := capnp.Client(session.resolverService.Capability())
	//forwarderMethod := service.NewForwarderMethod(m, &capResolverClient)
	forwarderMethod := service.NewForwarderMethod(m, session.capResolverService)
	return forwarderMethod
}

// ListCapabilities returns list of capabilities of all connected services sorted by service and capability names
func (session *GatewaySession) ListCapabilities(ctx context.Context) ([]resolver.CapabilityInfo, error) {
	return session.capList, nil
}

//
//// Login to the gateway
//// if no userauthn service is available then refuse
//// This sets the session clientID to the given ID when successful
//func (session *GatewaySession) Login(ctx context.Context, clientID, password string) (
//	authToken string, refreshToken string, err error) {
//
//	// need authn capability to login
//	userAuthn, err := session.getUserAuthn(ctx, clientID)
//	if err != nil {
//		return "", "", err
//	}
//	authToken, refreshToken, err = userAuthn.Login(ctx, password)
//
//	if err == nil {
//		logrus.Infof("Login of user '%s' successful.", clientID)
//		session.authType = hubapi.AuthTypeUser
//		session.clientID = clientID
//	} else {
//		session.authType = hubapi.AuthTypeUnauthenticated
//		err = fmt.Errorf("login of '%s' failed: %s", clientID, err)
//		logrus.Warning(err)
//	}
//	return authToken, refreshToken, err
//}
//
//// Ping capability
//func (session *GatewaySession) Ping(_ context.Context) (gateway.ClientInfo, error) {
//	logrus.Infof("Ping")
//	ci := gateway.ClientInfo{
//		ClientID: session.clientID,
//		AuthType: session.authType,
//	}
//	return ci, nil
//}
//
//// Refresh authentication tokens
//func (session *GatewaySession) Refresh(ctx context.Context, clientID string, oldRefreshToken string) (
//	authToken string, refreshToken string, err error) {
//
//	// use authn capability to refresh
//	userAuthn, err := session.getUserAuthn(ctx, clientID)
//	if err != nil {
//		return "", "", err
//	}
//	authToken, refreshToken, err = userAuthn.Refresh(ctx, oldRefreshToken)
//	return authToken, refreshToken, err
//}

// Release the connection to the resolver
func (session *GatewaySession) Release() {
	logrus.Infof("releasing session of client '%s'", session.clientID)
	session.capResolverService.Release()
}

// StartGatewaySession creates a new gateway session with the resolver to serve gateway requests.
// Use Release after the remote connection to the gateway is closed.
// This returns an error if connecting with the resolver fails.
// The user authentication is on loan to the session and should not be released.
//
//	resolverPath is the socket address for the resolver
//	clientID is the client ID
//	      use "" if the client is not authenticated and must use Login or Refresh
//	authType is the authentication type of the client, e.g. unauthenticated, service, device or user.
//	      using hubapi.AuthTypeUnauthenticated when not authenticated
//	clientConn is the TLS connection with the client. This will be closed on release
//	authnService optional authentication service. Intended for testing.
//func StartGatewaySession(
//	resolverPath string, clientID string, authType string, clientConn *tls.Conn,
//	authnService authn.IAuthnService) (*GatewaySession, error) {
//
//	session := &GatewaySession{
//		clientID:     clientID,
//		clientConn:   clientConn,
//		authType:     authType,
//		resolverPath: resolverPath,
//		authnService: authnService,
//	}
//	if clientConn != nil {
//
//	}
//	capClient, err := hubclient.ConnectWithCapnpUDS("", resolverPath)
//	if err != nil {
//		err = fmt.Errorf("unable to connect to the resolver socket at '%s': %s", resolverPath, err)
//		return nil, err
//	}
//	session.resolverService = capnpclient.NewResolverCapnpClient(capClient)
//
//	return session, err
//}

// NewGatewaySession creates a new gateway session
//
//	clientID is the authenticated ID of the client for this session
//	authType is the means of authentication
//	resCap is the resolver capability client
func NewGatewaySession(clientID string, authType string, resCap *capnpclient.ResolverCapnpClient) *GatewaySession {

	capList, _ := resCap.ListCapabilities(context.Background(), authType)

	s := &GatewaySession{
		capList:            capList,
		clientID:           clientID,
		authType:           authType,
		capResolverService: capnp.Client(resCap.Capability().AddRef()),
	}
	return s
}
