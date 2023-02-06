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
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
	"github.com/hiveot/hub/pkg/resolver/service"
)

// GatewaySession implements the IGatewaySession interface.
// A new instance is created by the capnp server for each incoming connection.
// This session is intended as a proxy for remote services to the local resolver.
type GatewaySession struct {

	// ID of the connected client
	clientID string

	// client connection
	clientConn *tls.Conn

	// type of the connected client
	clientType string

	// optional capability provider
	registeredProvider hubapi.CapProvider

	// optional provider capabilities
	registeredCapabilities []resolver.CapabilityInfo

	resolverPath    string
	resolverService *capnpclient.ResolverServiceCapnpClient
	resolverConn    net.Conn
	//authnSvc        authn.IAuthnService
	userAuthn authn.IUserAuthn // user auth capability obtained at login
}

// HandleUnknownMethod forwards the request to the resolver.
func (session *GatewaySession) HandleUnknownMethod(m capnp.Method) *server.Method {

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
	//logrus.Infof("clientTypes: %v", cstate.PeerCertificates[0].Subject.OrganizationalUnit)
	//logrus.Infof("handshake: %v", cstate.HandshakeComplete)
	//logrus.Infof("clientID='%s'", session.clientID)
	capList, err := session.resolverService.ListCapabilities(ctx, session.clientType)
	return capList, err
}

// Login to the gateway
// if no userauthn service is available then refuse
func (session *GatewaySession) Login(ctx context.Context, clientID, password string) (
	authToken string, refreshToken string, err error) {

	if session.userAuthn == nil {
		err = fmt.Errorf("sorry, authentication is not available")
		return
	}
	logrus.Infof("loginID=%s", clientID)
	authToken, refreshToken, err = session.userAuthn.Login(ctx, password)

	if err == nil {
		session.clientType = hubapi.ClientTypeUser
		session.clientID = clientID
	} else {
		session.clientType = hubapi.ClientTypeUnauthenticated
		err = fmt.Errorf("login of '%s' failed: %s", clientID, err)
		logrus.Warning(err)
	}
	return authToken, refreshToken, err
}

// Ping capability
func (session *GatewaySession) Ping(_ context.Context) (gateway.ClientInfo, error) {
	logrus.Infof("Ping")
	ci := gateway.ClientInfo{
		ClientID:   session.clientID,
		ClientType: session.clientType,
	}
	return ci, nil
}

// Refresh authentication tokens
func (session *GatewaySession) Refresh(ctx context.Context, oldRefreshToken string) (
	authToken string, refreshToken string, err error) {

	if session.userAuthn == nil {
		err = fmt.Errorf("not logged in")
		return "", "", err
	}
	authToken, refreshToken, err = session.userAuthn.Refresh(ctx, oldRefreshToken)
	return authToken, refreshToken, err
}

// Release the connection to the resolver
func (session *GatewaySession) Release() {
	logrus.Infof("releasing session of client '%s'", session.clientID)
	if session.resolverService != nil {
		session.resolverService.Release()
	}
	if session.registeredProvider.IsValid() {
		session.registeredProvider.Release()
	}
	if session.resolverConn != nil {
		_ = session.resolverConn.Close() // is this needed?
	}
	// do not release the userAuthn service as it belongs to the service, not the session
}

// RegisterCapabilities makes capabilities available to the hub.
// The session takes ownership of the provider and will release it on exit
//
//	clientID is the unique clientID of the capability provider
//	capInfo is the list with capabilities available through this provider
//	capProvider is the capnp capability provider callback interface used to obtain capabilities
//func (session *GatewaySession) RegisterCapabilities(_ context.Context,
//	clientID string, capInfo []resolver.CapabilityInfo, provider hubapi.CapProvider) error {
//
//	session.registeredCapabilities = capInfo
//	session.clientID = clientID
//	session.registeredProvider = provider
//	return nil
//}

// StartGatewaySession creates a new gateway session with the resolver to serve gateway requests.
// Use Release after the remote connection to the gateway is closed.
// This returns an error if connecting with the resolver fails.
// The user authentication is on loan to the session and should not be released.
//
//	userAuthn is the optional service to authenticate user requests. nil if user authentication is not available.
func StartGatewaySession(
	resolverPath string, userAuthn authn.IUserAuthn, clientConn *tls.Conn) (
	*GatewaySession, error) {

	ctx := context.Background()
	session := &GatewaySession{
		clientID:               "",
		clientConn:             clientConn,
		clientType:             hubapi.ClientTypeUnauthenticated,
		registeredProvider:     hubapi.CapProvider{},
		registeredCapabilities: nil,
		resolverPath:           resolverPath,
		userAuthn:              userAuthn,
	}
	resolverConn, err := net.Dial("unix", resolverPath)
	if err != nil {
		err = fmt.Errorf("unable to connect to the resolver socket at '%s': %s", resolverPath, err)
		return nil, err
	}

	session.resolverConn = resolverConn
	session.resolverService, err = capnpclient.NewResolverServiceCapnpClient(ctx, resolverConn)
	return session, err
}
