package listener

import (
	"context"
	"errors"
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

// RunService implements the boilerplate for running and shutting down a service using UDS sockets
//
//	serviceName used to create a UDS listening socket
//	socketFolder contains the location of service socket file
//	logsFolder to set logging to, or "" to not set logging output
//	startup is the method that starts the service and launches the capnp server
//	shutdown stops the service after the listener closes
func RunService(serviceName string, socketFolder string,
	startup func(ctx context.Context, lis net.Listener) error,
	shutdown func() error) {
	var err error

	// parse commandline and create server listening socket
	lis := CreateUDSServiceListener(socketFolder, serviceName)

	ctx := ExitOnSignal(context.Background(), func() {
		_ = lis.Close()
		_ = os.Remove(lis.Addr().String())
		err = shutdown()
	})

	// startup will wait until connection drops or  context is cancelled after signal is received
	// the result is the error indicating the reason
	err = startup(ctx, lis)

	if errors.Is(err, net.ErrClosed) {
		logrus.Warningf("%s service has shutdown gracefully", serviceName)
		os.Exit(0)
	} else {
		logrus.Errorf("%s service shutdown with error: %s", serviceName, err)
		os.Exit(-1)
	}
}
