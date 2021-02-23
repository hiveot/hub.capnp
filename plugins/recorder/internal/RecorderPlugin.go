package internal

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/pkg/messaging"
)

// RecorderConfig with logger plugin configuration
type RecorderConfig struct {
	gwConfig *lib.GatewayConfig
	Channels []string `yaml:"channels"`
}

// Recorder is a gateway plugin for recording channels
type Recorder struct {
	config *RecorderConfig
}

// Start the logging
func (rec *Recorder) Start() {
	logrus.Infof("Start: channels: %s", rec.config.Channels)
}

// Stop the logging
func (rec *Recorder) Stop() {
	logrus.Info("Stop ")

}

// NewRecorder creates a new recorder plugin instance
func NewRecorder(config *RecorderConfig) *Recorder {
	plugin := &Recorder{
		config: config,
	}
	// load the default channels
	if plugin.config.Channels == nil || len(plugin.config.Channels) == 0 {
		plugin.config.Channels = []string{messaging.TDChannelID, messaging.EventsChannelID, messaging.ActionChannelID}
	}
	return plugin
}
