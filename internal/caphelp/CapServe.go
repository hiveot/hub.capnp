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
		if err != nil {
			logrus.Errorf("CapServe accepting connections failed: %s", err)
			return err
		}
		go func() error {
			transport := rpc.NewStreamTransport(rwc)
			conn := rpc.NewConn(transport, &rpc.Options{
				BootstrapClient: client.AddRef(),
			})
			defer conn.Close()
			// Wait for connection to abort.
			select {
			case <-conn.Done():
				return nil
			case <-ctx.Done():
				return conn.Close()
			}
		}()
	}
}
