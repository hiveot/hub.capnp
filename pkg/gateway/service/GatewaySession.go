package service

import (
	"context"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
)

// GatewaySession implements the IGatewaySession interface.
// A new instance is created by the capnp server for each incoming connection.
// This session is intended as a proxy for remote services to the local resolver.
type GatewaySession struct {

	// ID of the connected client
	clientID string

	// type of the connected client
	clientType string

	// optional capability provider
	registeredProvider hubapi.CapProvider

	// optional provider capabilities
	registeredCapabilities []resolver.CapabilityInfo

	resolverPath    string
	resolverSession *capnpclient.ResolverSessionCapnpClient
	resolverConn    net.Conn
	authnSvc        authn.IAuthnService
	userAuthn       authn.IUserAuthn // user auth capability obtained at login
}

// Close the session
// This releases the capability connection if it exists
func (session *GatewaySession) Close() (err error) {
	logrus.Infof("closing session of client '%s'", session.clientID)
	if session.registeredProvider.IsValid() {
		session.registeredProvider.Release()
	}
	return nil
}

// GetCapability returns the capability with the given name, if available.
func (session *GatewaySession) GetCapability(ctx context.Context, clientID, clientType, capabilityName string, args []string) (
	capability capnp.Client, err error) {

	session.clientID = clientID
	session.clientType = clientType
	logrus.Infof("clientID='%s'", clientID)
	// TODO: authenticate and authorize the client - using middleware
	capability, err = session.resolverSession.GetCapability(ctx, clientID, clientType, capabilityName, args)
	// FIXME: error result from resolver is not returned
	if err == nil {
		err = capability.Resolve(ctx)
	}
	return capability, err
}

// ListCapabilities returns list of capabilities of all connected services sorted by service and capability names
func (session *GatewaySession) ListCapabilities(ctx context.Context) ([]resolver.CapabilityInfo, error) {
	capList := make([]resolver.CapabilityInfo, 0)

	//logrus.Infof("clientID='%s'", session.clientID)
	capList, err := session.resolverSession.ListCapabilities(ctx)
	return capList, err
}

// Login to the gateway
func (session *GatewaySession) Login(ctx context.Context, clientID, password string) (
	authToken string, refreshToken string, err error) {

	logrus.Infof("loginID=%s", clientID)

	// get the authentication capability from the resolver
	if err == nil {
		session.userAuthn = session.authnSvc.CapUserAuthn(ctx, clientID)
	}
	if err == nil {
		authToken, refreshToken, err = session.userAuthn.Login(ctx, password)
	}
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
	if session.resolverSession != nil {
		session.resolverSession.Release()
	}
	if session.registeredProvider.IsValid() {
		session.registeredProvider.Release()
	}
	if session.resolverConn != nil {
		_ = session.resolverConn.Close() // is this needed?
	}
}

// RegisterCapabilities makes capabilities available to the hub.
// The session takes ownership of the provider and will release it on exit
//
//	clientID is the unique clientID of the capability provider
//	capInfo is the list with capabilities available through this provider
//	capProvider is the capnp capability provider callback interface used to obtain capabilities
func (session *GatewaySession) RegisterCapabilities(_ context.Context,
	clientID string, capInfo []resolver.CapabilityInfo, provider hubapi.CapProvider) error {

	session.registeredCapabilities = capInfo
	session.clientID = clientID
	session.registeredProvider = provider
	return nil
}

// StartGatewaySession creates a new gateway session with the resolver to serve gateway requests.
// Use Release after the remote connection to the gateway is closed
// This returns an error if connecting with the resolver fails.
func StartGatewaySession(resolverPath string, authnSvc authn.IAuthnService) (*GatewaySession, error) {
	ctx := context.Background()
	session := &GatewaySession{
		clientID:               "",
		clientType:             hubapi.ClientTypeUnauthenticated,
		registeredProvider:     hubapi.CapProvider{},
		registeredCapabilities: nil,
		resolverPath:           resolverPath,
		authnSvc:               authnSvc,
	}
	resolverConn, err := net.Dial("unix", resolverPath)
	if err == nil {
		session.resolverConn = resolverConn
		session.resolverSession, err = capnpclient.NewResolverSessionCapnpClient(ctx, resolverConn)
	}
	return session, err
}
