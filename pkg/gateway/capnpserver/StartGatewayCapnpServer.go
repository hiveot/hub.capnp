package capnpserver

import (
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/gateway"
)

// StartGatewayCapnpServer starts listening for incoming capnp connections to the gateway.
// For each new connection new instances of the capnp server and gateway session are created.
// Each client therefore operates in its own session.
func StartGatewayCapnpServer(
	lis net.Listener, svc gateway.IGatewayService) error {

	logrus.Infof("listening on %s", lis.Addr().String())

	for {
		// Listen for incoming connections
		conn, err := lis.Accept()
		if err != nil {
			return err
		}

		logrus.Infof("New connection from remote client: %s. ID=%s",
			conn.RemoteAddr().String(), conn.RemoteAddr().String())

		// Each incoming connection is handled in a separate session.
		session := svc.OnIncomingConnection(conn)
		capsrv := NewGatewaySessionCapnpServer(session)

		// the bootstrap client session to pass to the remote client
		boot := hubapi.CapGatewaySession_ServerToClient(capsrv)
		//
		// the RPC connection takes ownership of the bootstrap interface and will release it when the connection
		// exits. Since this is a new instance for each connection there is no need to use AddRef.
		opts := rpc.Options{
			BootstrapClient: capnp.Client(boot), //.AddRef(),
		}
		// For each new incoming connection, create a new RPC transport connection that will serve incoming RPC requests
		transport := rpc.NewStreamTransport(conn)
		rpcConn := rpc.NewConn(transport, &opts)

		// The RPC connection is now established. Notify on disconnect
		go func() {
			<-rpcConn.Done()
			svc.OnConnectionClosed(conn, session)
		}()
	}
	return nil
}
