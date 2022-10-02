package listener

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

// ExitOnSignal starts a background process that invokes the shutdown callback,
// closes the context and listener and exits the service when a signal is received.
//  release is an optional application release function invoked before the shutdown.
//
func ExitOnSignal(ctx context.Context, listener net.Listener, release func()) {

	go func() {
		// catch all signals since not explicitly listing
		exitChannel := make(chan os.Signal, 1)

		//signal.Notify(exitChannel, syscall.SIGTERM|syscall.SIGHUP|syscall.SIGINT)
		signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM)

		sig := <-exitChannel
		logrus.Warningf("RECEIVED SIGNAL: %s", sig)

		if release != nil {
			release()
		}
		fmt.Println("Closing listening socket")
		listener.Close()
		ctx.Done()
		os.Exit(0)
	}()
}
