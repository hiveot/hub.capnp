package capnpserver

import (
	"context"
	"fmt"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/gateway/service"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

// GatewaySessionCapnpServer implements the capnp server of the gateway session.
// This implements the capnp hubapi.CapGatewaySession_server interface.
// Each incoming connection is served by its own session.
type GatewaySessionCapnpServer struct {
	session *service.GatewaySession // POGS service
}

func (capsrv *GatewaySessionCapnpServer) HandleUnknownMethod(m capnp.Method) *server.Method {
	// Just pass it on to the session that can add validation
	return capsrv.session.HandleUnknownMethod(m)
}

// ListCapabilities returns the aggregated list of capabilities from all connected services.
func (capsrv *GatewaySessionCapnpServer) ListCapabilities(
	ctx context.Context, call hubapi.CapGatewaySession_listCapabilities) (err error) {

	infoList, err := capsrv.session.ListCapabilities(ctx)
	resp, err2 := call.AllocResults()
	if err = err2; err == nil {
		infoListCapnp := capserializer.MarshalCapabilityInfoList(infoList)
		err = resp.SetInfoList(infoListCapnp)
	}
	return err
}

// Login to the session
func (capsrv *GatewaySessionCapnpServer) Login(
	ctx context.Context, call hubapi.CapGatewaySession_login) error {

	args := call.Args()
	loginID, _ := args.ClientID()
	password, _ := args.Password()
	authToken, refreshToken, err := capsrv.session.Login(ctx, loginID, password)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		if err == nil {
			err = res.SetAuthToken(authToken)
			_ = res.SetRefreshToken(refreshToken)
		}
	}
	return err
}

// Refresh authentication tokens
func (capsrv *GatewaySessionCapnpServer) Refresh(
	ctx context.Context, call hubapi.CapGatewaySession_refresh) error {
	args := call.Args()
	oldRefreshToken, _ := args.RefreshToken()
	authToken, refreshToken, err := capsrv.session.Refresh(ctx, oldRefreshToken)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetAuthToken(authToken)
		_ = res.SetRefreshToken(refreshToken)
	}
	return err
}

func (capsrv *GatewaySessionCapnpServer) Ping(
	ctx context.Context, call hubapi.CapGatewaySession_ping) error {

	response, err := capsrv.session.Ping(ctx)
	if err != nil {
		err = fmt.Errorf("ping somehow managed to fail")
		logrus.Error(err)
		return err
	}
	res, err := call.AllocResults()

	if err == nil {
		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		clientInfoCapnp, _ := hubapi.NewClientInfo(seg)
		_ = clientInfoCapnp.SetClientID(response.ClientID)
		_ = clientInfoCapnp.SetClientType(response.ClientType)
		err = res.SetReply(clientInfoCapnp)
	}
	return err
}

// Shutdown shut down the connection
func (capsrv *GatewaySessionCapnpServer) Shutdown() {
	capsrv.session.Release()
}

// NewGatewaySessionCapnpServer creates a capnp server session to serve a new connection
// session is unfortunately the gateway session implementation, not its interface, because
// access to handling unknown method is needed.
func NewGatewaySessionCapnpServer(session *service.GatewaySession) *GatewaySessionCapnpServer {

	srv := &GatewaySessionCapnpServer{
		session: session,
	}
	return srv
}

// StartGatewaySessionCapnpServer instantiates a new session using the given
// gateway session.
//func StartGatewaySessionCapnpServer(
//	session gateway.IGatewaySession, conn net.Conn) (*rpc.Conn, error) {
//
//	srv := &GatewaySessionCapnpServer{
//		session: session,
//	}
//	srv.ResolverSessionCapnpServer = *capnpserver.NewResolverSessionCapnpServer(session)
//
//	// the RPC connection takes ownership of the bootstrap interface and will release it
//	// when the connection exits. No need to use AddRef.
//	main := hubapi.CapGatewaySession_ServerToClient(srv)
//	opts := rpc.Options{
//		BootstrapClient: capnp.Client(main),
//	}
//	// For each new incoming connection, create a new RPC transport connection that will serve incoming RPC requests
//	transport := rpc.NewStreamTransport(conn)
//	rpcConn := rpc.NewConn(transport, &opts)
//	srv.rpcConn = rpcConn
//	return rpcConn
//}
