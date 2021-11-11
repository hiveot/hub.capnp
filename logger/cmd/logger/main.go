package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/proc"
	"github.com/wostzone/hub/logger/internal"
)

func main() {
	svc := internal.NewLoggerService()
	hubConfig, err := config.LoadAllConfig(os.Args, "", internal.PluginID, &svc.Config)
	if err != nil {
		logrus.Errorf("ERROR: Start aborted due to error")
		os.Exit(1)
	}
	err = svc.Start(hubConfig)
	if err != nil {
		logrus.Errorf("Logger: Failed to start")
		os.Exit(1)
	}
	proc.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
