package main

import (
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/proc"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/logger/internal"
)

func main() {
	svc := internal.NewLoggerService()
	hubConfig, err := config.LoadAllConfig(os.Args, "", internal.PluginID, &svc.Config)
	if err != nil {
		logrus.Errorf("ERROR: Start aborted due to error")
		os.Exit(1)
	}
	logging.SetLogging(hubConfig.Loglevel, hubConfig.LogFile)
	err = svc.Start(hubConfig)
	if err != nil {
		logrus.Errorf("Logger: Failed to start")
		os.Exit(1)
	}
	proc.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
