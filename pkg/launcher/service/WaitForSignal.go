package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

// WaitForSignal waits until a TERM or INT signal is received
// Intended for use by hub services to run until the app is done
func WaitForSignal() {

	// catch all signals since not explicitly listing
	exitChannel := make(chan os.Signal, 1)

	//signal.Notify(exitChannel, syscall.SIGTERM|syscall.SIGHUP|syscall.SIGINT)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM)

	sig := <-exitChannel
	logrus.Warningf("RECEIVED SIGNAL: %s", sig)
	fmt.Println()
	fmt.Println(sig)
}
