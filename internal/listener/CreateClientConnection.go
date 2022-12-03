package listener

import (
	"fmt"
	"net"
	"path/filepath"
	"time"
)

// CreateClientConnection returns a local client connection for the given service
//
// The service itself must listen on the unix domain socket for the service following the
// convention: {runFolder}/{serviceName}.socket
//
//	runFolder is the folder containing sockets
//	serviceName is the name of the service to connect to
func CreateClientConnection(runFolder, serviceName string) (net.Conn, error) {
	svcAddress := filepath.Join(runFolder, serviceName+".socket")
	conn, err := net.DialTimeout("unix", svcAddress, time.Second)
	if err != nil {
		err = fmt.Errorf("Unable to connect to service socket '%s'. Is the service running?\n  Error: %s", svcAddress, err)
	}
	return conn, err
}
