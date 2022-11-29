package caphelp

import (
	"context"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/listener"
)

// CapServe starts serving requests using the given listener and capnp capability client
// The function ends when the connection is closed or the context is done
// capnpService is the capnp server obtained with Xyz_ServerToClient that implements the
// knownMethods defined in the capnp schema for interface Xyz.
func CapServe(parentCtx context.Context, serviceName string, lis net.Listener, capnpService capnp.Client) error {
	ctx := listener.ExitOnSignal(parentCtx, serviceName, nil)
	//var contextDone = false

	// listen.Accept() does not use context, so this is a workaround
	go func() {
		select {
		case <-ctx.Done():
			logrus.Infof("%s: Context of exitonsignal cancelled. Closing connection.", serviceName)
			_ = lis.Close()
		}
	}()

	// Listen for new connections
	for {
		rwc, err := lis.Accept()
		if ctx.Err() != nil {
			logrus.Infof("ctx.err for '%s': %s ", serviceName, ctx.Err())
			break
		}
		if err != nil {
			logrus.Errorf("%s: listener closed outside of context: %s. No use continuing.", serviceName, err)
			return err
		}
		// for each new incoming connection, create a client instance to be returned to the remote
		// peer (eg,the remote user).
		// multiple incoming connections are supported so run them in a go-routine.
		// when the context closes, all client connections will close.
		go func() {
			connID := getConnectionID(rwc)
			logrus.Infof("%s: New connection from remote client: %s. ID=%s",
				serviceName, rwc.RemoteAddr().String(), connID)

			transport := rpc.NewStreamTransport(rwc)
			conn := rpc.NewConn(transport, &rpc.Options{
				BootstrapClient: capnpService.AddRef(),
			})
			//defer conn.Close()
			// Wait for connection to abort or context to cancel
			select {
			case <-conn.Done():
				logrus.Infof("%s: Remote client connection closed. ID=%s", serviceName, connID)
				return
			case <-ctx.Done():
				logrus.Infof("%s: Service context cancelled. Closing client connection '%s'", serviceName, connID)
				_ = conn.Close()
				return
			}
		}()
	}
	logrus.Infof("%s: capserve ended", serviceName)
	return nil
}

// getConnectionID returns the ID of the unix domain or TCP socket connection.
// used to pair incoming and closing connections in the logs
// This returns 0 if the connection is not unix or tcp, or is closed.
func getConnectionID(rwc net.Conn) string {
	udc, found := rwc.(*net.UnixConn)
	if found {
		fd, _ := udc.File()
		fdName := fd.Name()
		fdfd := fd.Fd()
		idText := fmt.Sprintf("%s [%d]", fdName, fdfd)
		return idText
	}
	tcp, found := rwc.(*net.TCPConn)
	if found {
		ra := tcp.RemoteAddr()
		fd, _ := tcp.File()
		fdfd := fd.Fd()
		idText := fmt.Sprintf("%s [%d]", ra, fdfd)
		return idText
	}
	return "closed"
}
