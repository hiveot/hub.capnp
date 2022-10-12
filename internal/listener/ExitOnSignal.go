package listener

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// ExitOnSignal starts a background process and closes the context when a SIGINT or SIGTERM is received.
// if a release function is provided, it is invoked first.
// This returns a child context which is cancelled on receiving a signal
func ExitOnSignal(ctx context.Context, serviceName string, release func()) context.Context {
	exitCtx, cancelFn := context.WithCancel(ctx)
	go func() {
		// catch all signals since not explicitly listing
		exitChannel := make(chan os.Signal, 1)

		//signal.Notify(exitChannel, syscall.SIGTERM|syscall.SIGHUP|syscall.SIGINT)
		signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM)

		sig := <-exitChannel
		logrus.Warningf("RECEIVED SIGNAL for service '%s': %s", serviceName, sig)

		if release != nil {
			release()
		}
		// cancel the context. This should invoke Done()
		cancelFn()
		time.Sleep(time.Second)

	}()
	return exitCtx
}
