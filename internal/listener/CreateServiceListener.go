package listener

import (
	"net"
	"path/filepath"
	"syscall"

	"github.com/sirupsen/logrus"
)

// CreateServiceListener creates a local TCP or Unix domain socket listener with the given service name
// for use by capnp, http or grpc servers.
//
// By convention, HiveOT Hub Services listen on Unix Sockets on address {runFolder}/{serviceName}.socket.
// Remote clients must use the gateway to connect to the service.
//
//  runFolder provides the Hub sockets folder
//  serviceName provides the name of the socket
// This returns a listening socket for accepting incoming connections
func CreateServiceListener(runFolder, serviceName string) net.Listener {
	var address = "localhost"
	var unixSocket = filepath.Join(runFolder, serviceName+".socket")

	// listen on tcp port or unix domain socket
	address = unixSocket
	// remove stale handle
	_ = syscall.Unlink(address)
	listener, err := net.Listen("unix", address)

	if err != nil {
		logrus.Fatalf("failed to create a listener on %s: %v", address, err)
	}
	logrus.Infof("Listening on %v", listener.Addr())
	return listener
}
