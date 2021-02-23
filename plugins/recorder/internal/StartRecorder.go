package internal

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
)

const pluginID = "recorder"

// StartRecorder reads the configuration, parses the commandline and start the plugin
// The recoder will start recording the channels configured in recorder.yaml
func StartRecorder(appFolder string) (*Recorder, error) {
	loggerConfig := &RecorderConfig{}
	gatewayConfig, err := lib.SetupConfig(appFolder, pluginID, loggerConfig)
	loggerConfig.gwConfig = gatewayConfig

	if err != nil {
		return nil, err
	}
	rec := NewRecorder(loggerConfig)
	rec.Start()
	return rec, nil
}

// StopRecorder stops the running recorder
func StopRecorder() {
	logrus.Warningf("Stopping %s", pluginID)
}
