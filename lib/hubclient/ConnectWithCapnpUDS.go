package hubclient

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"context"
	"fmt"
	"net"
	"time"
)

// ConnectWithCapnpUDS creates a connection to a service over Unix Domain Sockets
// using the convention that connection address = {socketFolder}/{serviceName}.socket
// If no serviceName is given, socketFolder is expected to contain the full socket path.
// To close the connection the client must be released
func ConnectWithCapnpUDS(
	serviceName, socketFolder string) (capClient capnp.Client, err error) {

	socketPath := fmt.Sprintf("%s/%s.socket", socketFolder, serviceName)
	timeout := time.Second * 3
	if serviceName == "" {
		socketPath = socketFolder
	}
	conn, err := net.DialTimeout("unix", socketPath, timeout)
	tp := rpc.NewStreamTransport(conn)
	rpcConn := rpc.NewConn(tp, nil)
	capClient = rpcConn.Bootstrap(context.Background())
	return capClient, err
}
