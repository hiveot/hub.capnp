package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/plugins/recorder/internal"
)

var recorderConfig = &internal.RecorderConfig{}

func main() {
	gatewayConfig, err := lib.SetupConfig("", internal.RecorderPluginID, recorderConfig)

	svc := internal.NewRecorderService()
	err = svc.Start(gatewayConfig, recorderConfig)
	if err != nil {
		logrus.Errorf("recorder: Failed to start")
		os.Exit(1)
	}
	lib.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
