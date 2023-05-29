package capnpserver

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"
	"context"
	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/gateway/service"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

// GatewaySessionCapnpServer implements the capnp server of the gateway session.
// This implements the capnp hubapi.CapProvider interface.
// Each incoming connection is served by its own session.
type GatewaySessionCapnpServer struct {
	session *service.GatewaySession // POGS service
}

// HandleUnknownMethod is a hook into capnp server that is invoked when a requested
// method is not found.
// This passes the request to the session which forwards it to the resolver service.
func (capsrv *GatewaySessionCapnpServer) HandleUnknownMethod(m capnp.Method) *server.Method {
	// Just pass it on to the session that can add validation of the method
	return capsrv.session.HandleUnknownMethod(m)
}

// ListCapabilities returns the aggregated list of capabilities from all connected services.
func (capsrv *GatewaySessionCapnpServer) ListCapabilities(
	ctx context.Context, call hubapi.CapProvider_listCapabilities) (err error) {

	infoList, err := capsrv.session.ListCapabilities(ctx)
	resp, err2 := call.AllocResults()
	if err = err2; err == nil {
		infoListCapnp := capserializer.MarshalCapabilityInfoList(infoList)
		err = resp.SetInfoList(infoListCapnp)
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
