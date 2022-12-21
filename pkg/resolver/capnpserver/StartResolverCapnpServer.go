package capnpserver

import (
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/resolver"
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

		// these capnp and resolver sessions are going to handle the connection
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
		_ = rpcConn
		// the RPC connection is now established
		// notify on disconnect
		go func() {
			<-rpcConn.Done()
			//boot.Release() // probably not needed ???
			svc.OnConnectionClosed(conn, session)
		}()

		//--- determine if the remote bootstrap client is a provider ---
		// if the service is using it.
		//remoteBoot := rpcConn.Bootstrap(context.Background())
		//err := remoteBoot.Resolve(ctx)
		//if err != nil {
		//	logrus.Warningf("Unable to resolve client: %s", err)
		//	return
		//}
		//// we can't tell if the remote client supports the resolver API so just make the call to find out
		//capProvider := hubapi.CapProvider(remoteBoot)
		//remoteBoot.State().Metadata.Put("onincomingconnection", "testing")
		//connectionID := capProvider.String()
		//method, release := capProvider.ListCapabilities(ctx, nil)
		//defer release()
		//resp, err := method.Struct()
		//if err != nil {
		//	logrus.Infof("New connection from '%s'. This is not a provider.", connectionID)
		//	svc.connectClients.Add(1)
		//	go func() {
		//		<-rpcConn.Done()
		//		svc.connectClients.Add(-1)
		//		remaining := svc.connectClients.Load()
		//		logrus.Infof("Client connection '%s' closed. %d remaining", connectionID, remaining)
		//	}()
		//} else {
		//	// liftoff, this is a client that has capabilities to offer. Store them and keep the connection.
		//	// add the remote client capabilities and keep track of the connection
		//	capInfoListCapnp, err := resp.InfoList()
		//	if err == nil {
		//	}
		//	capInfoList := capserializer.UnmarshalCapabilyInfoList(capInfoListCapnp)
		//	// The resolver will be released on stop.
		//	err = svc.RegisterCapabilities(ctx, connectionID, capInfoList, capProvider)
		//	// cleanup
		//	go func() {
		//		<-rpcConn.Done()
		//		logrus.Infof("Provider connection '%s' closed. %d providers remaining",
		//			connectionID, len(svc.connectedProviders))
		//		svc.removeService(connectionID)
		//	}()
		//
		//}
		//return err
	}
}
