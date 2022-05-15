package internal

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/wostzone/hub/lib/client/pkg/certsclient"
	"github.com/wostzone/hub/lib/client/pkg/mqttbinding"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
	"github.com/wostzone/hub/lib/client/pkg/thing"
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
	Config     WostLoggerConfig
	hubConfig  *config.HubConfig
	mqttClient *mqttclient.MqttClient
	loggers    map[string]*os.File // map of thing ID to logfile
	isRunning  bool                // not intended for concurrent use
}

// handleMessage receives and records a topic message
// thingID is the ID of the thing from its TD
// msgType the operation, e.g. read, write, event, action
// subject of the operation, event name, properties, action name
func (wlog *LoggerService) logToFile(thingID string, msgType string, subject string, payload []byte) {
	logrus.Infof("Received message of type '%s/%s' about Thing %s", msgType, subject, thingID)
	// var err error

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
	//parsedMsg := make(map[string]interface{})
	var parsedMsg interface{}
	_ = json.Unmarshal(payload, &parsedMsg)

	logMsg := make(map[string]interface{})
	logMsg["receivedAt"] = time.Now().Format("2006-01-02T15:04:05.000-0700")
	logMsg["payload"] = parsedMsg
	logMsg["thingID"] = thingID
	logMsg["msgType"] = msgType
	logMsg["subject"] = subject
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
	thingID := thing.CreatePublisherID(wlog.hubConfig.Zone, "hub", wlog.Config.ClientID, deviceType)
	logrus.Infof("Publishing this service TD %s", thingID)
	thingTD := thing.CreateTD(thingID, PluginID, deviceType)
	thingTD.UpdateTitleDescription("Simple Hub message logging", "This service logs hub messages to file")
	thingTD.UpdateProperty("logsFolder", &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type:  vocab.WoTDataTypeString,
			Title: "Directory where to store the log files",
		},
	})
	eThing := mqttbinding.CreateExposedThing(wlog.Config.ClientID, thingTD, wlog.mqttClient)
	eThing.Expose()
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

	// connect to the message bus to receive messages
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
	wlog.mqttClient = mqttclient.NewMqttClient(wlog.Config.ClientID, caCert, 0)
	err = wlog.mqttClient.ConnectWithClientCert(hostPort, pluginCert)
	if err != nil {
		return err
	}

	if wlog.Config.ThingIDs == nil || len(wlog.Config.ThingIDs) == 0 {
		// log everything
		wlog.mqttClient.Subscribe("#", func(address string, payload []byte) {
			thingID, msgType, subject := mqttbinding.SplitTopic(address)
			wlog.logToFile(thingID, msgType, subject, payload)
		})
	} else {
		for _, thingID := range wlog.Config.ThingIDs {
			topic := mqttbinding.CreateTopic(thingID, "#")
			wlog.mqttClient.Subscribe(topic,
				func(address string, payload []byte) {
					cbThingID, msgType, subject := mqttbinding.SplitTopic(address)
					wlog.logToFile(cbThingID, msgType, subject, payload)
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
	// remove subscriptions before closing loggers
	for _, thingID := range wlog.Config.ThingIDs {
		topic := mqttbinding.CreateTopic(thingID, "#")
		wlog.mqttClient.Unsubscribe(topic)
	}
	for _, logger := range wlog.loggers {
		// logger.Out.(*os.File).Close()
		logger.Close()
	}
	wlog.loggers = nil
	wlog.mqttClient.Close()
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
