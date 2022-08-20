package listener

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
)

// CreateServiceListener creates a TCP or Unix domain socket listener with the given service name
// for use by http or grpc server.
//
// This parses the commandline for options '-a address' or '-u unixsocket' to listen on
// The default socket is /tmp/{serviceName}.sock
// In case of error this exits with Fatal.
//
// By default this listens on the unix domain socket /tmp/serviceName.sock
// Any additional commandline option flags must be set before invoking this method.
// Returns the listening socket.
func CreateServiceListener(serviceName string) net.Listener {
	var address string = ""
	var unixSocket string = "/tmp/" + serviceName + ".sock"
	flag.Usage = func() {
		fmt.Printf("Usage: %s [-a address | -u /path/to/unixdomainsocket]\n", os.Args[0])
		flag.PrintDefaults()
		//"thingstore [-p port|domainsocket]"
	}
	flag.StringVar(&unixSocket, "u", unixSocket, "GRPC listening unix domain socket")
	flag.StringVar(&address, "a", address, "GRPC listening [address]:port")
	flag.Parse()

	// listen on tcp port or unix domain socket
	network := "unix"
	if address != "" {
		network = "tcp"
	} else {
		address = unixSocket
		// remove stale handle
		// TODO: send a terminate message to the socket in case it is used
		_ = syscall.Unlink(address)
	}
	listener, err := net.Listen(network, address)

	if err != nil {
		log.Fatalf("failed to create a %s listener on %s: %v", network, address, err)
	}
	log.Printf("Listening on %v", listener.Addr())
	return listener
}
