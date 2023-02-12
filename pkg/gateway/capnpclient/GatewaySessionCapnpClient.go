package capnpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

type GatewaySessionCapnpClient struct {
	connection *rpc.Conn                // connection to capnp server
	capability hubapi.CapGatewaySession // capnp client of the gateway session
}

// ListCapabilities lists the available capabilities of the service
// Returns a list of capabilities that can be obtained through the service
func (cl *GatewaySessionCapnpClient) ListCapabilities(
	ctx context.Context) (infoList []resolver.CapabilityInfo, err error) {

	infoList = make([]resolver.CapabilityInfo, 0)
	method, release := cl.capability.ListCapabilities(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		infoListCapnp, err2 := resp.InfoList()
		if err = err2; err == nil {
			infoList = capserializer.UnmarshalCapabilyInfoList(infoListCapnp)
		}
	}
	return infoList, err
}

// Login to the gateway
func (cl *GatewaySessionCapnpClient) Login(ctx context.Context,
	clientID string, password string) (authToken, refreshToken string, err error) {

	method, release := cl.capability.Login(ctx,
		func(params hubapi.CapGatewaySession_login_Params) error {
			err = params.SetClientID(clientID)
			_ = params.SetPassword(password)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		authToken, err = resp.AuthToken()
		refreshToken, _ = resp.RefreshToken()
	}
	return authToken, refreshToken, err
}

// Ping performs a ping test
func (cl *GatewaySessionCapnpClient) Ping(
	ctx context.Context) (clientInfo gateway.ClientInfo, err error) {

	method, release := cl.capability.Ping(ctx, nil)
	defer release()

	resp, err := method.Struct()
	if err == nil {
		clInfoCapnp, err2 := resp.Reply()
		err = err2
		clientInfo.ClientID, _ = clInfoCapnp.ClientID()
		clientInfo.AuthType, _ = clInfoCapnp.AuthType()
	}
	return clientInfo, err
}

// Refresh auth tokens
func (cl *GatewaySessionCapnpClient) Refresh(ctx context.Context,
	clientID string, oldRefreshToken string) (authToken, refreshToken string, err error) {

	method, release := cl.capability.Refresh(ctx,
		func(params hubapi.CapGatewaySession_refresh_Params) error {
			err = params.SetRefreshToken(oldRefreshToken)
			_ = params.SetClientID(clientID)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		authToken, err = resp.AuthToken()
		refreshToken, _ = resp.RefreshToken()
	}
	return authToken, refreshToken, err
}

// RegisterCapabilities registers a service's capabilities along with the CapProvider
//func (cl *GatewaySessionCapnpClient) RegisterCapabilities(ctx context.Context,
//	serviceID string, capInfoList []resolver.CapabilityInfo,
//	capProvider hubapi.CapProvider) (err error) {
//
//	capInfoListCapnp := capserializer.MarshalCapabilityInfoList(capInfoList)
//	method, release := cl.capability.RegisterCapabilities(ctx,
//		func(params hubapi.CapResolverSession_registerCapabilities_Params) error {
//			err = params.SetCapInfo(capInfoListCapnp)
//			_ = params.SetServiceID(serviceID)
//			_ = params.SetProvider(capProvider.AddRef()) // don't forget AddRef
//			return err
//		})
//	defer release()
//	_, err = method.Struct()
//	return err
//}

// Release the client
func (cl *GatewaySessionCapnpClient) Release() {
	cl.capability.Release()
	if cl.connection != nil {
		err := cl.connection.Close()
		if err != nil {
			logrus.Error(err)
		}
	}
}

// ConnectToGateway is a helper that starts a new connection with the gateway
// over TLS.
// Users should call Release when done. This will close the connection and any
// capabilities obtained from the resolver.
//
//	fullUrl of the server: eg tcp://server:port, or wss://server:port/ws
//	clientCert is the TLS client certificate for mutual authentication. Use nil to connect
//			   as an unauthenticated client.
//	caCert is the server's CA certificate to verify that the gateway service is valid.
//			   Use nil to not verify the server's certificate.
//				network is either "unix" or "tcp". Default "" uses "tcp"
//				address is the UDS or TCP address:port of the gateway
//
// This returns a client for a gateway session
func ConnectToGateway(fullUrl string,
	clientCert *tls.Certificate, caCert *x509.Certificate) (
	gatewayClient gateway.IGatewaySession, err error) {

	rpcCon, hubClient, err := hubclient.ConnectToHubClient(fullUrl, clientCert, caCert)

	capGatewaySession := hubapi.CapGatewaySession(hubClient)

	cl := &GatewaySessionCapnpClient{
		connection: rpcCon,
		capability: capGatewaySession,
	}
	return cl, err
}

// NewGatewaySessionFromCapnpCapability returns a POGS wrapper around the gateway capnp instance
func NewGatewaySessionFromCapnpCapability(capability hubapi.CapGatewaySession) gateway.IGatewaySession {
	gws := GatewaySessionCapnpClient{capability: capability}
	return &gws
}
