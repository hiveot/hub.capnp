// package main for the mosquitto manager
package main

import (
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/proc"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/mosquittomgr/internal"
)

// Main entry to WoST plugin for managing Mosquitto
// This setup the configuration from file and commandline parameters and launches the service
func main() {
	svc := internal.NewMosquittoManager()
	hubConfig, err := config.LoadAllConfig(os.Args, "", internal.PluginID, &svc.Config)
	if err != nil {
		logrus.Errorf("Mosquittomgr: Start aborted due to commandline error")
		os.Exit(1)
	}

	err = svc.Start(hubConfig)
	if err != nil {
		logrus.Errorf("Mosquittomgr: Failed to start: %s", err)
		os.Exit(1)
	}
	proc.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
