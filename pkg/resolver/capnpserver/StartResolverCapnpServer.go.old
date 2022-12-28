package capnpserver

import (
	"context"
	"net"
	"os"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/service"
)

// StartResolverCapnpServer starts a new resolver capnp server for incoming connections.
// For each incoming connection a new session is created that has its own bootstrap instance so
// that disconnects can close the session and remove any subscriptions.
// A new boostrap instance is needed to identify the session of incoming messages.
func StartResolverCapnpServer(
	lis net.Listener, svc resolver.IResolverService) error {

	for {
		// Accept incoming connections
		conn, err := lis.Accept()
		if err != nil {
			return err
		}

		logrus.Infof("New connection from remote client: %s", conn.RemoteAddr().String())

		// a capnp and resolver sessions are going to handle the connection
		session := svc.OnIncomingConnection(conn)
		capsrv := NewResolverSessionCapnpServer(session)

		// the bootstrap client session to pass to the remote client
		boot := hubapi.CapResolverSession_ServerToClient(capsrv)

		// the RPC connection takes ownership of the bootstrap interface and will release it when the connection
		// exits. Since this is a new instance for each connection there is no need to use AddRef.
		opts := rpc.Options{
			BootstrapClient: capnp.Client(boot), //.AddRef(),
		}
		// For each new incoming connection, create a new RPC transport connection that will serve incoming RPC requests
		transport := rpc.NewStreamTransport(conn)
		rpcConn := rpc.NewConn(transport, &opts)

		// the RPC connection is now established. Notify on disconnect
		go func() {
			<-rpcConn.Done()
			svc.OnConnectionClosed(conn, session)
		}()

	}
}

// StartResolver is a helper for starting the resolver service and its capnp server in the background.
// This returns a stop function or an error if start fails.
func StartResolver(socketPath string) (stopFn func(), err error) {
	if socketPath == "" {
		socketPath = resolver.DefaultResolverPath
	}
	_ = os.RemoveAll(socketPath)
	svc := service.NewResolverService()
	err = svc.Start(context.Background())
	if err != nil {
		logrus.Panicf("Failed to start with socket dir %s", socketPath)
	}

	logrus.Infof("listening on %s", socketPath)
	lis, err := net.Listen("unix", socketPath)
	if err == nil {
		go StartResolverCapnpServer(lis, svc)
	}
	return func() {
		_ = svc.Stop()
		_ = lis.Close()
	}, err
}
