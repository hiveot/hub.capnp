package internal

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/pkg/messaging"
)

// RecorderConfig with logger plugin configuration
type RecorderConfig struct {
	Channels []string `yaml:"channels"`
}

// Recorder is a gateway plugin for recording channels
type Recorder struct {
	config      RecorderConfig
	gwConfig    *lib.GatewayConfig
	messenger   messaging.IGatewayMessenger
	fileHandles map[string]*os.File
}

// RecorderPluginID is the default ID of the recorder plugin
const RecorderPluginID = "recorder"

// handleChannelMessage receives and records a channel message
func (rec *Recorder) handleChannelMessage(channel string, message []byte) {
	logrus.Infof("handleChannelMessage: Received message on channel %s: %s", channel, message)
	fileHandle := rec.fileHandles[channel]
	if fileHandle != nil {
		sender := ""
		timeStamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
		// timeStamp := time.Now().Format(time.RFC3339Nano)
		maxLen := len(message)
		if maxLen > 40 {
			maxLen = 40
		}
		line := fmt.Sprintf("[%s] %s %s: %s", timeStamp, sender, channel, message[:maxLen])
		n, err := fileHandle.WriteString(line + "\n")
		_ = n
		if err != nil {
			logrus.Errorf("handleChannelMessage: Unable to record channel '%s': %s", channel, err)
		}
	}
}

// StartRecordChannel setup recording of a channel
func (rec *Recorder) StartRecordChannel(channel string, messenger messaging.IGatewayMessenger) {
	logsFolder := path.Dir(rec.gwConfig.Logging.LogFile)
	filename := path.Join(logsFolder, channel+".txt")
	fileHandle, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)

	if err != nil {
		logrus.Errorf("StartRecordChannel: Unable to open file '%s' for writing: %s. Channel '%s' ignored", filename, err, channel)
		return
	}
	rec.fileHandles[channel] = fileHandle
	messenger.Subscribe(channel, rec.handleChannelMessage)
}

// Start connects, subscribe and start the recording
func (rec *Recorder) Start(gwConfig *lib.GatewayConfig, recConfig *RecorderConfig) error {
	var err error
	rec.config = *recConfig
	rec.gwConfig = gwConfig
	rec.messenger, err = messaging.StartGatewayMessenger(RecorderPluginID, gwConfig)

	// messaging.NewGatewayConnection(gwConfig.Messenger.Protocol)
	// load the default channels if config doesn't have any
	if rec.config.Channels == nil || len(rec.config.Channels) == 0 {
		rec.config.Channels = []string{messaging.TDChannelID, messaging.EventsChannelID, messaging.ActionChannelID}
	}
	for _, channel := range rec.config.Channels {
		rec.StartRecordChannel(channel, rec.messenger)
	}

	logrus.Infof("Start: channels: %s", rec.config.Channels)
	return err
}

// Stop the logging
func (rec *Recorder) Stop() {
	logrus.Info("Recorder Stop: Stopping recorder service")
	for _, channel := range rec.config.Channels {
		rec.messenger.Unsubscribe(channel)
	}
	for _, fileHandle := range rec.fileHandles {
		fileHandle.Close()
	}
	rec.fileHandles = make(map[string]*os.File)

}

// NewRecorderService creates a new recorder service instance
func NewRecorderService() *Recorder {
	rec := &Recorder{
		fileHandles: make(map[string]*os.File),
	}
	return rec
}
