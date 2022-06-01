package internal

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/consumedthing"
	"github.com/wostzone/wost-go/pkg/exposedthing"
	"github.com/wostzone/wost-go/pkg/mqttclient"
	"github.com/wostzone/wost-go/pkg/thing"
	"github.com/wostzone/wost-go/pkg/vocab"

	"github.com/sirupsen/logrus"
)

// PluginID is the default ID of the WoST Logger plugin
const PluginID = "logger"

// LoggerServiceConfig with logger plugin configuration
// map of topic -> file
type LoggerServiceConfig struct {
	// unique service ID of logger instance. Default is plugin ID.
	ClientID string `yaml:"clientID"`
	// Expose this service with a Thing. Default is false
	ExposeService bool `yaml:"exposeService"`
	// folder to use for logging. Required.
	LogFolder string `yaml:"logFolder"`
	// Mqtt broker address. Required
	MqttAddress string `yaml:"mqttAddress"`
	// Mqtt broker port for certificate auth. Default is config.DefaultMqttPortCert
	MqttPortCert int `yaml:"mqttPortCert"`
	// thing IDs to log. Default is all of them.
	ThingIDs []string `yaml:"thingIDs"`
}

// LoggerService is a hub plugin for recording messages to the hub
// By default it logs messages by ThingID, eg each Thing has a log file
type LoggerService struct {
	// CA certificate to verify the mqtt broker connection
	caCert *x509.Certificate

	// Client certificate to authenticate with the mqtt broker
	clientCert *tls.Certificate

	// The logger configuration
	config LoggerServiceConfig

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
		filePath := path.Join(ls.config.LogFolder, thingID+".log")

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
	thingID := thing.CreatePublisherID("", "", ls.config.ClientID, deviceType)

	logrus.Infof("Publishing this service TD %s", thingID)
	serviceTD := thing.CreateTD(thingID, PluginID, deviceType)
	serviceTD.UpdateTitleDescription("Simple Hub message logging", "This service logs hub messages to file")
	serviceTD.UpdateProperty("logsFolder", &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Type:  vocab.WoTDataTypeString,
			Title: "Directory where to store the log files",
		},
	})
	ls.etFactory.Expose(ls.config.ClientID, serviceTD)
	//eThing := mqttbinding.CreateExposedThing(ls.config.ClientID, thingTD, ls.mqttClient)
	//eThing.Expose()
}

// Start connects, subscribe and start the recording
func (ls *LoggerService) Start() error {
	var err error

	// check for required configuration
	if ls.config.LogFolder == "" || ls.config.MqttAddress == "" {
		err := errors.New("missing required configuration. Unable to continue")
		logrus.Error(err)
		return err
	}

	// create the log folder if needed
	err = os.MkdirAll(ls.config.LogFolder, 0700)
	if err != nil {
		err2 := errors.New("Failed creating the log destination folder: " + err.Error())
		logrus.Error(err2)
		return err2
	}

	// connect to the message bus to receive messages to be logged
	hostPort := fmt.Sprintf("%s:%d", ls.config.MqttAddress, ls.config.MqttPortCert)
	// the mqtt client is used for subscribing to messages to log
	ls.mqttClient = mqttclient.NewMqttClient(ls.config.ClientID, ls.caCert, 0)
	err = ls.mqttClient.ConnectWithClientCert(hostPort, ls.clientCert)
	if err != nil {
		return err
	}

	if ls.config.ThingIDs == nil || len(ls.config.ThingIDs) == 0 {
		// log all things
		topic := consumedthing.CreateTopic("+", "#")
		ls.mqttClient.Subscribe(topic, func(address string, payload []byte) {
			thingID, msgType, subject := consumedthing.SplitTopic(address)
			ls.logToFile(thingID, msgType, subject, payload)
		})
	} else {
		for _, thingID := range ls.config.ThingIDs {
			topic := consumedthing.CreateTopic(thingID, "#")
			ls.mqttClient.Subscribe(topic,
				func(address string, payload []byte) {
					cbThingID, msgType, subject := consumedthing.SplitTopic(address)
					ls.logToFile(cbThingID, msgType, subject, payload)
				})
		}
	}

	// Expose the service if configured
	if ls.config.ExposeService {
		ls.etFactory = exposedthing.CreateExposedThingFactory(ls.config.ClientID, ls.clientCert, ls.caCert)
		err = ls.etFactory.Connect(ls.config.MqttAddress, ls.config.MqttPortCert)
		ls.ExposeService()
	}

	logrus.Infof("Started logger of %d topics", len(ls.config.ThingIDs))
	ls.isRunning = true
	return err
}

// Stop the service
func (ls *LoggerService) Stop() {
	if !ls.isRunning {
		return
	}
	logrus.Info("Stopping logging service")
	// remove subscriptions before closing loggers
	for _, thingID := range ls.config.ThingIDs {
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
func NewLoggerService(
	loggerConfig LoggerServiceConfig,
	clientCert *tls.Certificate,
	caCert *x509.Certificate) *LoggerService {

	// set defaults
	if loggerConfig.ClientID == "" {
		loggerConfig.ClientID = PluginID
	}
	if loggerConfig.MqttPortCert == 0 {
		loggerConfig.MqttPortCert = config.DefaultMqttPortCert
	}

	svc := &LoggerService{
		config:     loggerConfig,
		caCert:     caCert,
		clientCert: clientCert,
		loggers:    make(map[string]*os.File),
	}
	return svc
}
