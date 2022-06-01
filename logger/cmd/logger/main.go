package main

import (
	"os"

	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/proc"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/hub/logger/internal"
)

func main() {
	svcConfig := internal.LoggerServiceConfig{}
	hubConfig, err := config.LoadAllConfig(os.Args, "", internal.PluginID, &svcConfig)
	if err != nil {
		logrus.Errorf("ERROR: Start aborted due to error")
		os.Exit(1)
	}
	logging.SetLogging(hubConfig.Loglevel, hubConfig.LogFile)
	if svcConfig.MqttAddress == "" {
		svcConfig.MqttAddress = hubConfig.Address
	}
	if svcConfig.MqttPortCert == 0 {
		svcConfig.MqttPortCert = hubConfig.MqttPortCert
	}
	if svcConfig.LogFolder == "" {
		svcConfig.LogFolder = hubConfig.LogFolder
	}
	svc := internal.NewLoggerService(svcConfig, hubConfig.PluginCert, hubConfig.CaCert)
	err = svc.Start()
	if err != nil {
		logrus.Errorf("Logger: Failed to start")
		os.Exit(1)
	}
	proc.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
