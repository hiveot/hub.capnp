package capnpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

type GatewaySessionCapnpClient struct {
	connection *rpc.Conn                // connection to capnp server
	capability hubapi.CapGatewaySession // capnp client of the gateway session
}

// GetCapability obtains the capability with the given name.
// The caller must release the capability when done.
func (cl *GatewaySessionCapnpClient) GetCapability(ctx context.Context,
	clientID string, clientType string, capabilityName string, args []string) (
	capabilityRef capnp.Client, err error) {

	method, release := cl.capability.GetCapability(ctx,
		func(params hubapi.CapResolverSession_getCapability_Params) error {
			_ = params.SetClientID(clientID)
			_ = params.SetClientType(clientType)
			_ = params.SetCapName(capabilityName)
			if args != nil {
				err = params.SetArgs(caphelp.MarshalStringList(args))
			}
			return err
		})
	defer release()
	// return a future. Caller must release
	//capability = method.Cap().AddRef()

	// Just return the actual capability instead of a future, so the error is obtained if it isn't available.
	// Would be nice to return the future but this is an infrequent call anyways.
	resp, err := method.Struct()
	if err == nil {
		capability := resp.Capability().AddRef()
		capabilityRef = capability
	}
	return capabilityRef, err
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
		clientInfo.ClientType, _ = clInfoCapnp.ClientType()
	}
	return clientInfo, err
}

// Refresh auth tokens
func (cl *GatewaySessionCapnpClient) Refresh(ctx context.Context,
	oldRefreshToken string) (authToken, refreshToken string, err error) {

	method, release := cl.capability.Refresh(ctx,
		func(params hubapi.CapGatewaySession_refresh_Params) error {
			err = params.SetRefreshToken(oldRefreshToken)
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
func (cl *GatewaySessionCapnpClient) RegisterCapabilities(ctx context.Context,
	serviceID string, capInfoList []resolver.CapabilityInfo,
	capProvider hubapi.CapProvider) (err error) {

	capInfoListCapnp := capserializer.MarshalCapabilityInfoList(capInfoList)
	method, release := cl.capability.RegisterCapabilities(ctx,
		func(params hubapi.CapResolverSession_registerCapabilities_Params) error {
			err = params.SetCapInfo(capInfoListCapnp)
			_ = params.SetServiceID(serviceID)
			_ = params.SetProvider(capProvider.AddRef()) // don't forget AddRef
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

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

// NewGatewaySessionCapnpClient create a new gateway client for obtaining capabilities.
// Intended for remote clients such as IoT devices, services or users to connect to the
// Hub's gateway. A connection must be established first.
//
//	conn is the network connection to use.
func NewGatewaySessionCapnpClient(ctx context.Context,
	conn net.Conn) (cl *GatewaySessionCapnpClient, err error) {

	transport := rpc.NewStreamTransport(conn)
	rpcConn := rpc.NewConn(transport, nil)
	capGatewaySession := hubapi.CapGatewaySession(rpcConn.Bootstrap(ctx))

	cl = &GatewaySessionCapnpClient{
		connection: rpcConn,
		capability: capGatewaySession,
	}
	return cl, nil
}

// ConnectToGatewayTLS is a helper that starts a new connection with the gateway
// over TLS.
// Users should call Release when done. This will close the connection and any
// capabilities obtained from the resolver.
//
//	 clientCert is the TLS client certificate for mutual authentication. Use nil to connect
//	   as an unauthenticated client.
//	 caCert is the server's CA certificate to verify that the gateway service is valid.
//	   Use nil to not verify the server's certificate.
//		network is either "unix" or "tcp". Default "" uses "tcp"
//		address is the UDS or TCP address:port of the gateway
//
// This returns a client for a gateway session
func ConnectToGatewayTLS(network, address string,
	clientCert *tls.Certificate, caCert *x509.Certificate) (

	gatewayClient gateway.IGatewaySession, err error) {

	if address == "" {
		panic("need a gateway address")
	} else if network == "" {
		network = "tcp"
	}
	// create the TLS connection for use by the RPC
	clConn, err := listener.CreateTLSClientConnection(network, address, clientCert, caCert)
	if err != nil {
		logrus.Panic(err)
	}
	ctx := context.Background()
	cl, err := NewGatewaySessionCapnpClient(ctx, clConn)
	return cl, err
}