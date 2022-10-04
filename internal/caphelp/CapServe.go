package caphelp

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/listener"
)

// CapServe starts serving requests using the given listener and capnp capability client
// The function ends when the connection is closed or the context is done
// This creates a new client instance for each incoming connection
func CapServe(ctx context.Context, lis net.Listener, client capnp.Client) error {
	listener.ExitOnSignal(ctx, lis, nil)

	// Listen for calls
	for {
		rwc, err := lis.Accept()
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			logrus.Errorf("CapServe accepting connections failed: %s", err)
			return err
		}
		// for each new incoming connection, create a client instance to be returned to the remote
		// peer (eg,the remote user).
		// multiple incoming connections are supported so run them in a go-routine.
		// when the context closes, all client connections will close.
		go func() {
			transport := rpc.NewStreamTransport(rwc)
			conn := rpc.NewConn(transport, &rpc.Options{
				BootstrapClient: client.AddRef(),
			})
			defer conn.Close()
			// Wait for connection to abort or context to cancel
			select {
			case <-conn.Done():
				return
			case <-ctx.Done():
				logrus.Infof("Context cancelled. Closing connection.")
				_ = conn.Close()
				return
			}
		}()
	}
	logrus.Infof("capserve ended")
	return nil
}
