package listener

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
)

// RunService implements the boilerplate for running and shutting down a service using UDS sockets
//
//	serviceName used to create a UDS listening socket
//	socketPath contains the location of service listening socket file
//	startup is the method that starts the service and launches the capnp server
//	shutdown stops the service after the listener closes
func RunService(serviceName string, socketPath string,
	startup func(ctx context.Context, lis net.Listener) error,
	shutdown func() error) {
	var err error

	// parse commandline and create server listening socket
	//lis := CreateUDSServiceListener(socketFolder, serviceName)
	// remove stale handle
	_ = syscall.Unlink(socketPath)
	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		err2 := fmt.Errorf("failed to create a listener on %s: %v", socketPath, err)
		logrus.Fatal(err2)
	}
	logrus.Infof("Listening on %v", socketPath)

	ctx := ExitOnSignal(context.Background(), func() {
		logrus.Infof("Stopping service '%s'", serviceName)
		_ = lis.Close()
		_ = os.Remove(lis.Addr().String())
	})

	// startup will wait until connection drops or  context is cancelled after signal is received
	// the result is the error indicating the reason
	err = startup(ctx, lis)

	logrus.Infof("Service '%s' shutting down", serviceName)

	err = shutdown()

	if errors.Is(err, net.ErrClosed) {
		logrus.Infof("Service '%s' has stopped gracefully", serviceName)
		os.Exit(0)
	} else if err != nil {
		logrus.Errorf("Service '%s' shutdown with error: %s", serviceName, err)
		os.Exit(-1)
	}
	logrus.Infof("Service '%s' has shutdown with no errors", serviceName)
}
