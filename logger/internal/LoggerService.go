package internal

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/certs"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/td"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

// PluginID is the default ID of the WoST Logger plugin
const PluginID = "logger"

// WostLoggerConfig with logger plugin configuration
// map of topic -> file
type WostLoggerConfig struct {
	ClientID   string   `yaml:"clientID"`   // custom unique client ID of logger instance
	PublishTD  bool     `yaml:"publishTD"`  // publish the TD of this service
	LogsFolder string   `yaml:"logsFolder"` // folder to use for logging
	ThingIDs   []string `yaml:"thingIDs"`   // thing IDs to log
}

// LoggerService is a hub plugin for recording messages to the hub
// By default it logs messages by ThingID, eg each Thing has a log file
type LoggerService struct {
	Config        WostLoggerConfig
	hubConfig     *config.HubConfig
	hubConnection *mqttclient.MqttHubClient
	loggers       map[string]*os.File // map of thing ID to logfile
	isRunning     bool                // not intended for concurrent use
}

// handleMessage receives and records a topic message
func (wlog *LoggerService) logToFile(thingID string, msgType string, payload []byte, sender string) {
	logrus.Infof("Received message of type '%s' about Thing %s", msgType, thingID)
	// var err error
	_ = sender

	if wlog.loggers == nil {
		logrus.Errorf("logToFile called after logger has stopped")
		return
	}

	logger := wlog.loggers[thingID]
	if logger == nil {
		logsFolder := wlog.Config.LogsFolder
		filePath := path.Join(logsFolder, thingID+".log")

		// 	TimestampFormat: "2006-01-02T15:04:05.000-0700",
		fileHandle, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
		if err != nil {
			logrus.Errorf("Unable to open logfile for Thing: %s", filePath)
			return
		} else {
			logrus.Infof("Created logfile for Thing: %s", filePath)
		}
		wlog.loggers[thingID] = fileHandle
		logger = fileHandle
	}
	parsedMsg := make(map[string]interface{})
	logMsg := make(map[string]interface{})
	logMsg["receivedAt"] = time.Now().Format("2006-01-02T15:04:05.000-0700")
	logMsg["sender"] = ""
	logMsg["payload"] = parsedMsg
	logMsg["thingID"] = thingID
	logMsg["msgType"] = msgType
	_ = json.Unmarshal(payload, &parsedMsg)
	pretty, _ := json.MarshalIndent(logMsg, " ", "  ")
	prettyStr := string(pretty) + ",\n"
	_, _ = logger.WriteString(prettyStr)
}

// PublishServiceTD publishes the Thing Description of the logger service itself
func (wlog *LoggerService) PublishServiceTD() {
	if !wlog.Config.PublishTD {
		return
	}
	deviceType := vocab.DeviceTypeService
	thingID := td.CreatePublisherThingID(wlog.hubConfig.Zone, "hub", wlog.Config.ClientID, deviceType)
	logrus.Infof("Publishing this service TD %s", thingID)
	thingTD := td.CreateTD(thingID, PluginID, deviceType)
	// Include the logging folder as a property
	prop := td.CreateProperty("Logging Folder", "Directory where to store the log files", vocab.PropertyTypeAttr)
	td.SetPropertyDataTypeString(prop, 0, 0)
	//
	td.AddTDProperty(thingTD, "logsFolder", prop)
	wlog.hubConnection.PublishTD(thingID, thingTD)
	td.SetThingDescription(thingTD, "Simple Hub message logging", "This service logs hub messages to file")
}

// Start connects, subscribe and start the recording
func (wlog *LoggerService) Start(hubConfig *config.HubConfig) error {
	var err error
	var pluginCert *tls.Certificate
	// wlog.loggers = make(map[string]*logrus.Logger)
	wlog.loggers = make(map[string]*os.File)
	wlog.hubConfig = hubConfig

	// verify the logging folder exists
	if wlog.Config.LogsFolder == "" {
		// default location is hubConfig log folder
		wlog.Config.LogsFolder = wlog.hubConfig.LogsFolder
	} else if !path.IsAbs(wlog.Config.LogsFolder) {
		wlog.Config.LogsFolder = path.Join(hubConfig.AppFolder, wlog.Config.LogsFolder)
	}
	_, err = os.Stat(wlog.Config.LogsFolder)
	if err != nil {
		logrus.Errorf("Start: Logging folder does not exist: %s. Setup error: %s", wlog.Config.LogsFolder, err)
		return err
	}

	// connect the the message bus to receive messages
	caCertPath := path.Join(hubConfig.CertsFolder, config.DefaultCaCertFile)
	caCert, err := certs.LoadX509CertFromPEM(caCertPath)
	if err == nil {
		pluginCert, err = certs.LoadTLSCertFromPEM(
			path.Join(hubConfig.CertsFolder, config.DefaultPluginCertFile),
			path.Join(hubConfig.CertsFolder, config.DefaultPluginKeyFile),
		)
	}
	if err != nil {
		logrus.Errorf("Start: Error loading certificate: %s", err)
		return err
	}
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortCert)
	wlog.hubConnection = mqttclient.NewMqttHubClient(wlog.Config.ClientID, caCert)
	err = wlog.hubConnection.ConnectWithClientCert(hostPort, pluginCert)
	if err != nil {
		return err
	}

	if wlog.Config.ThingIDs == nil || len(wlog.Config.ThingIDs) == 0 {
		// log everything
		wlog.hubConnection.Subscribe("", func(thingID string, msgType string, payload []byte, senderID string) {
			wlog.logToFile(thingID, msgType, payload, senderID)
		})
	} else {
		for _, thingID := range wlog.Config.ThingIDs {
			wlog.hubConnection.Subscribe(thingID,
				func(evThingID string, msgType string, payload []byte, senderID string) {
					wlog.logToFile(evThingID, msgType, payload, senderID)
				})
		}
	}

	// publish the logger service thing
	wlog.PublishServiceTD()

	logrus.Infof("Started logger of %d topics", len(wlog.Config.ThingIDs))
	wlog.isRunning = true
	return err
}

// Stop the logging
func (wlog *LoggerService) Stop() {
	if !wlog.isRunning {
		return
	}
	logrus.Info("Stopping logging service")
	if len(wlog.Config.ThingIDs) == 0 {
		wlog.hubConnection.Unsubscribe("")
	} else {
		for _, thingID := range wlog.Config.ThingIDs {
			wlog.hubConnection.Unsubscribe(thingID)
		}
	}
	for _, logger := range wlog.loggers {
		// logger.Out.(*os.File).Close()
		logger.Close()
	}
	wlog.loggers = nil
	wlog.hubConnection.Close()
	wlog.isRunning = false
}

// NewLoggerService returns a new instance of the logger service
func NewLoggerService() *LoggerService {
	svc := &LoggerService{
		Config: WostLoggerConfig{
			ClientID:   PluginID,
			PublishTD:  false,
			LogsFolder: "",
		},
	}
	return svc
}
