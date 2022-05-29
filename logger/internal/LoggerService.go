package internal

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/consumedthing"
	"github.com/wostzone/wost-go/pkg/exposedthing"
	"github.com/wostzone/wost-go/pkg/mqttclient"
	"github.com/wostzone/wost-go/pkg/thing"
	"github.com/wostzone/wost-go/pkg/vocab"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
)

// PluginID is the default ID of the WoST Logger plugin
const PluginID = "logger"

// WostLoggerConfig with logger plugin configuration
// map of topic -> file
type WostLoggerConfig struct {
	ClientID      string   `yaml:"clientID"`      // custom unique client ID of logger instance
	ExposeService bool     `yaml:"exposeService"` // Expose this service with a Thing
	LogsFolder    string   `yaml:"logsFolder"`    // folder to use for logging
	ThingIDs      []string `yaml:"thingIDs"`      // thing IDs to log
}

// LoggerService is a hub plugin for recording messages to the hub
// By default it logs messages by ThingID, eg each Thing has a log file
type LoggerService struct {
	// The logger configuration
	Config WostLoggerConfig

	// Shared configuration of the hub
	hubConfig *config.HubConfig

	// Map of thing ID to logfile
	loggers map[string]*os.File

	// service is running (not intended for concurrent use)
	isRunning bool

	// etFactory to expose this service if Config.ExposeService is set
	etFactory *exposedthing.ExposedThingFactory

	// mqtt client for receiving messages to be logged
	mqttClient *mqttclient.MqttClient
}

// handleMessage receives and records a topic message
//  thingID is the ID of the thing from its TD
//  msgType the operation, e.g. read, write, event, action
//  subject of the operation, event name, properties, action name
//  payload to log
func (ls *LoggerService) logToFile(thingID string, msgType string, subject string, payload []byte) {
	logrus.Infof("Received message of type '%s/%s' about Thing %s", msgType, subject, thingID)
	// var err error

	if ls.loggers == nil {
		logrus.Errorf("logToFile called after logger has stopped")
		return
	}

	logger := ls.loggers[thingID]
	if logger == nil {
		logsFolder := ls.Config.LogsFolder
		filePath := path.Join(logsFolder, thingID+".log")

		// 	TimestampFormat: "2006-01-02T15:04:05.000-0700",
		fileHandle, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
		if err != nil {
			logrus.Errorf("Unable to open logfile for Thing: %s", filePath)
			return
		} else {
			logrus.Infof("Created logfile for Thing: %s", filePath)
		}
		ls.loggers[thingID] = fileHandle
		logger = fileHandle
	}
	//parsedMsg := make(map[string]interface{})
	var parsedMsg interface{}
	_ = json.Unmarshal(payload, &parsedMsg)

	logMsg := make(map[string]interface{})
	logMsg["receivedAt"] = time.Now().Format(vocab.TimeFormat)
	logMsg["payload"] = parsedMsg
	logMsg["thingID"] = thingID
	logMsg["msgType"] = msgType
	logMsg["subject"] = subject
	pretty, _ := json.MarshalIndent(logMsg, " ", "  ")
	prettyStr := string(pretty) + ",\n"
	_, _ = logger.WriteString(prettyStr)
}

// ExposeService exposes the logger service itself as a Thing
// This will not do anything if Config.ExposeService was disabled at start
//
// This create a TD for this service and publish it using the exposed thing factory.
//
func (ls *LoggerService) ExposeService() {
	if ls.etFactory == nil {
		return
	}
	// create the TD of this service
	deviceType := vocab.DeviceTypeService
	thingID := thing.CreatePublisherID(ls.hubConfig.Zone, "hub", ls.Config.ClientID, deviceType)

	logrus.Infof("Publishing this service TD %s", thingID)
	serviceTD := thing.CreateTD(thingID, PluginID, deviceType)
	serviceTD.UpdateTitleDescription("Simple Hub message logging", "This service logs hub messages to file")
	serviceTD.UpdateProperty("logsFolder", &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type:  vocab.WoTDataTypeString,
			Title: "Directory where to store the log files",
		},
	})
	ls.etFactory.Expose(ls.Config.ClientID, serviceTD)
	//eThing := mqttbinding.CreateExposedThing(ls.Config.ClientID, thingTD, ls.mqttClient)
	//eThing.Expose()
}

// Start connects, subscribe and start the recording
func (ls *LoggerService) Start(hubConfig *config.HubConfig) error {
	var err error
	var pluginCert *tls.Certificate
	// ls.loggers = make(map[string]*logrus.Logger)
	ls.loggers = make(map[string]*os.File)
	ls.hubConfig = hubConfig

	// verify the logging folder exists
	if ls.Config.LogsFolder == "" {
		// default location is hubConfig log folder
		ls.Config.LogsFolder = ls.hubConfig.LogFolder
	} else if !path.IsAbs(ls.Config.LogsFolder) {
		ls.Config.LogsFolder = path.Join(hubConfig.HomeFolder, ls.Config.LogsFolder)
	}
	_, err = os.Stat(ls.Config.LogsFolder)
	if err != nil {
		logrus.Errorf("Logging folder '%s' does not exist. Setup error: %s",
			ls.Config.LogsFolder, err)
		return err
	}

	// connect to the message bus to receive messages to be logged
	caCertPath := path.Join(hubConfig.CertsFolder, config.DefaultCaCertFile)
	caCert, err := certsclient.LoadX509CertFromPEM(caCertPath)
	if err == nil {
		pluginCert, err = certsclient.LoadTLSCertFromPEM(
			path.Join(hubConfig.CertsFolder, config.DefaultPluginCertFile),
			path.Join(hubConfig.CertsFolder, config.DefaultPluginKeyFile),
		)
	}
	if err != nil {
		logrus.Errorf("Start: Error loading certificate: %s", err)
		return err
	}
	hostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortCert)
	// the mqtt client is used for subscribing to messages to log
	ls.mqttClient = mqttclient.NewMqttClient(ls.Config.ClientID, caCert, 0)
	err = ls.mqttClient.ConnectWithClientCert(hostPort, pluginCert)
	if err != nil {
		return err
	}

	if ls.Config.ThingIDs == nil || len(ls.Config.ThingIDs) == 0 {
		// log everything
		ls.mqttClient.Subscribe("#", func(address string, payload []byte) {
			thingID, msgType, subject := consumedthing.SplitTopic(address)
			ls.logToFile(thingID, msgType, subject, payload)
		})
	} else {
		for _, thingID := range ls.Config.ThingIDs {
			topic := consumedthing.CreateTopic(thingID, "#")
			ls.mqttClient.Subscribe(topic,
				func(address string, payload []byte) {
					cbThingID, msgType, subject := consumedthing.SplitTopic(address)
					ls.logToFile(cbThingID, msgType, subject, payload)
				})
		}
	}

	// Expose the service if configured
	if ls.Config.ExposeService {
		ls.etFactory = exposedthing.CreateExposedThingFactory(ls.Config.ClientID, pluginCert, caCert)
		ls.etFactory.Connect(hubConfig.Address, hubConfig.MqttPortCert)
		ls.ExposeService()
	}

	logrus.Infof("Started logger of %d topics", len(ls.Config.ThingIDs))
	ls.isRunning = true
	return err
}

// Stop the logging
func (ls *LoggerService) Stop() {
	if !ls.isRunning {
		return
	}
	logrus.Info("Stopping logging service")
	// remove subscriptions before closing loggers
	for _, thingID := range ls.Config.ThingIDs {
		topic := consumedthing.CreateTopic(thingID, "#")
		ls.mqttClient.Unsubscribe(topic)
	}
	for _, logger := range ls.loggers {
		// logger.Out.(*os.File).Close()
		logger.Close()
	}
	ls.loggers = nil
	ls.mqttClient.Disconnect()
	ls.isRunning = false
}

// NewLoggerService returns a new instance of the logger service
func NewLoggerService() *LoggerService {
	svc := &LoggerService{
		Config: WostLoggerConfig{
			ClientID:      PluginID,
			ExposeService: false,
			LogsFolder:    "",
		},
	}
	return svc
}
