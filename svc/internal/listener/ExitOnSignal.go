package listener

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

// ExitOnSignal invokes the shutdown callback, closes the listener and exits the service when a signal is received
// This captures notify in a separate goroutine.
func ExitOnSignal(listener net.Listener, shutdown func()) {

	go func() {
		// catch all signals since not explicitly listing
		exitChannel := make(chan os.Signal, 1)

		//signal.Notify(exitChannel, syscall.SIGTERM|syscall.SIGHUP|syscall.SIGINT)
		signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM)

		sig := <-exitChannel
		logrus.Warningf("RECEIVED SIGNAL: %s", sig)

		if shutdown != nil {
			shutdown()
		}
		fmt.Println("Closing listening socket")
		listener.Close()
		os.Exit(0)
	}()
}
