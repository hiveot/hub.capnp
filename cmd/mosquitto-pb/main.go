// package main for the protocol binding
package main

import (
	"os"

	"github.com/sirupsen/logrus"
	mosquittopb "github.com/wostzone/hub/core/mosquitto-pb"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// Main entry to WoST protocol adapter for managing Mosquitto
// This setup the configuration from file and commandline parameters and launches the service
func main() {
	svc := mosquittopb.NewMosquittoManager()
	hubConfig, err := hubconfig.LoadCommandlineConfig("", mosquittopb.PluginID, &svc.Config)
	if err != nil {
		logrus.Errorf("ERROR: Start aborted due to error")
		os.Exit(1)
	}

	err = svc.Start(hubConfig)
	if err != nil {
		logrus.Errorf("Logger: Failed to start: %s", err)
		os.Exit(1)
	}
	hubclient.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
