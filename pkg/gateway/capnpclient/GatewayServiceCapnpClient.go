package capnpclient

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"context"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/api/go/hubapi"
)

type GatewayServiceCapnpClient struct {
	connection *rpc.Conn                // rpc connection to capnp gateway server
	capability hubapi.CapGatewayService // capnp client of the gateway service
}

func (cl *GatewayServiceCapnpClient) AuthNoAuth(clientID string) (sessionToken string) {

	method, release := cl.capability.AuthNoAuth(context.Background(),
		func(params hubapi.CapGatewayService_authNoAuth_Params) error {
			err := params.SetClientID(clientID)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		sessionToken, err = resp.SessionToken()
	}
	return sessionToken
}

func (cl *GatewayServiceCapnpClient) AuthProxy(clientID string, clientCertPEM string) (sessionToken string) {
	method, release := cl.capability.AuthProxy(context.Background(),
		func(params hubapi.CapGatewayService_authProxy_Params) error {
			err := params.SetClientID(clientID)
			_ = params.SetClientCertPEM(clientCertPEM)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		sessionToken, err = resp.SessionToken()
	}
	return sessionToken
}

func (cl *GatewayServiceCapnpClient) AuthRefresh(clientID string, oldSessionToken string) (sessionToken string) {
	method, release := cl.capability.AuthRefresh(context.Background(),
		func(params hubapi.CapGatewayService_authRefresh_Params) error {
			err := params.SetClientID(clientID)
			_ = params.SetSessionToken(oldSessionToken)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		sessionToken, err = resp.SessionToken()
	}
	return sessionToken
}

func (cl *GatewayServiceCapnpClient) AuthWithCert() (sessionToken string) {

	method, release := cl.capability.AuthWithCert(context.Background(), nil)

	defer release()
	resp, err := method.Struct()
	if err == nil {
		sessionToken, _ = resp.SessionToken()
	}
	return sessionToken
}

// AuthWithPassword login to the gateway using password
func (cl *GatewayServiceCapnpClient) AuthWithPassword(
	clientID string, password string) (sessionToken string) {

	method, release := cl.capability.AuthWithPassword(context.Background(),
		func(params hubapi.CapGatewayService_authWithPassword_Params) error {
			err := params.SetClientID(clientID)
			_ = params.SetPassword(password)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		sessionToken, err = resp.SessionToken()
	}
	return sessionToken
}

func (cl *GatewayServiceCapnpClient) NewSession(
	sessionToken string) (session resolver.ICapProvider, err error) {

	method, release := cl.capability.NewSession(context.Background(),
		func(params hubapi.CapGatewayService_newSession_Params) error {
			err := params.SetSessionToken(sessionToken)
			return err
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capProvider := capnp.Client(resp.Session())
		session = NewGatewaySessionCapnpClient(capProvider)
	}
	return session, err
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
func (cl *GatewayServiceCapnpClient) Release() {
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
//	fullUrl of the server: eg tcp://server:port, or wss://server:port/ws, or "" for auto discovery
//	searchTimeSec of autodiscovery or 0 for default
//	clientCert is the TLS client certificate for mutual authentication. Use nil to connect
//				   as an unauthenticated client.
//	caCert is the server's CA certificate to verify that the gateway service is valid.
//				   Use nil to not verify the server's certificate.
//					network is either "unix" or "tcp". Default "" uses "tcp"
//					address is the UDS or TCP address:port of the gateway
//
// This returns a client for a gateway session
//func ConnectToGateway(fullUrl string, searchTimeSec int,
//	clientCert *tls.Certificate, caCert *x509.Certificate) (
//	gatewayClient gateway.IGatewaySession, err error) {
//
//	rpcCon, hubClient, err := hubclient.ConnectWithCapnp(fullUrl, searchTimeSec, clientCert, caCert)
//
//	capGatewaySession := hubapi.CapGatewaySession(hubClient)
//
//	cl := &GatewaySessionCapnpClient{
//		connection: rpcCon,
//		capability: capGatewaySession,
//	}
//	return cl, err
//}

// NewGatewayServiceCapnpClient returns a POGS wrapper around the gateway capnp instance
func NewGatewayServiceCapnpClient(capClient capnp.Client) *GatewayServiceCapnpClient {
	capGateway := hubapi.CapGatewayService(capClient)
	gws := GatewayServiceCapnpClient{
		capability: capGateway,
		connection: nil,
	}
	return &gws
}
