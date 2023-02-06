package capnpserver

import (
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"capnproto.org/go/capnp/v3/rpc/transport"
	"capnproto.org/go/capnp/v3/server"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/pkg/gateway/service"
)

// StartGatewayCapnpServer starts listening for incoming capnp connections to the gateway.
// For each new connection new instances of the capnp server and gateway session are created.
// Each client therefore operates in its own session.
//
//	svc is the gateway service to serve
//	lis is the socket listener
//	useWS set to use a websocket transport
func StartGatewayCapnpServer(svc *service.GatewayService, lis net.Listener, useWS bool) error {
	useWSText := ""
	if useWS {
		useWSText = " using websockets"
	}
	logrus.Infof("listening on %s%s", lis.Addr(), useWSText)

	// Each incoming connection is handled in a separate session.
	// This handler will create a new capnp client and a gateway session object.
	onConnect := func(conn net.Conn, tp transport.Transport) {
		session := svc.OnIncomingConnection(conn)
		if session == nil {
			conn.Close()
			return
		}
		capSession := NewGatewaySessionCapnpServer(session)

		// Instead of using ServerToClient, use server.New to be able to
		// add the 'handleUnknownMethod' hook.
		//boot := hubapi.CapGatewaySession_ServerToClient(capsrv)
		c, _ := hubapi.CapGatewaySession_Server(capSession).(server.Shutdowner)
		methods := hubapi.CapGatewaySession_Methods(nil, capSession)
		clientHook := server.New(methods, capSession, c)
		clientHook.HandleUnknownMethod = capSession.HandleUnknownMethod
		resClient := capnp.NewClient(clientHook)
		boot := hubapi.CapGatewaySession(resClient)

		opts := rpc.Options{
			BootstrapClient: capnp.Client(boot),
		}
		// Each connection gets a new RPC transport that will serve incoming RPC requests
		// transport := rpc.NewStreamTransport(conn)
		rpcConn := rpc.NewConn(tp, &opts)
		go func() {
			<-rpcConn.Done()
			logrus.Infof("Connection from '%s' closed", conn.RemoteAddr())
		}()
	}

	if useWS {
		return listener.ServeWSCB(lis, onConnect)
	} else {
		return listener.ServeCB(lis, onConnect)
	}
}

// StartGatewayCapnpServer starts listening for incoming capnp connections to the gateway.
// For each new connection new instances of the capnp server and gateway session are created.
// Each client therefore operates in its own session.
//
//	svc is the gateway service to serve
//	lis is the socket listener
//	useWS set to use a websocket transport
// func StartGatewayCapnpServerOrg(svc *service.GatewayService, lis net.Listener, useWS bool) error {
// 	useWSText := ""
// 	if useWS {
// 		useWSText = " using websockets"
// 	}
// 	logrus.Infof("listening on %s%s", lis.Addr(), useWSText)

// 	for {
// 		// Listen for incoming connections
// 		conn, err := lis.Accept()
// 		if err != nil {
// 			return err
// 		}

// 		// Each incoming connection is handled in a separate session.
// 		session := svc.OnIncomingConnection(conn)
// 		if session != nil {
// 			var tp transport.Transport

// 			capsrv := NewGatewaySessionCapnpServer(session)

// 			// Instead of using ServerToClient, use server.New to be able to
// 			// add the 'handleUnknownMethod' hook.
// 			//boot := hubapi.CapGatewaySession_ServerToClient(capsrv)
// 			c, _ := hubapi.CapGatewaySession_Server(capsrv).(server.Shutdowner)
// 			methods := hubapi.CapGatewaySession_Methods(nil, capsrv)
// 			clientHook := server.New(methods, capsrv, c)
// 			clientHook.HandleUnknownMethod = capsrv.HandleUnknownMethod

// 			resClient := capnp.NewClient(clientHook)
// 			boot := hubapi.CapGatewaySession(resClient)

// 			if useWS {
// 				codec := websocketcapnp.NewCodec(conn, true)
// 				tp = transport.New(codec)
// 			} else {
// 				tp = rpc.NewStreamTransport(conn)
// 			}

// 			listener.ServeTransport(
// 				gateway.ServiceName,
// 				tp,
// 				capnp.Client(boot),
// 				nil,
// 				func(c *rpc.Conn) { svc.OnConnectionClosed(conn, session) },
// 			)
// 		}
// 	}
// }
