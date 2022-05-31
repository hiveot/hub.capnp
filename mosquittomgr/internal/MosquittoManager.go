package internal

import (
	"os"
	"os/exec"

	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/mqttclient"

	"github.com/sirupsen/logrus"
)

const PluginID = "mosquittomgr"

const DefaultConfFile = "mosquitto.conf"
const DefaultTemplateFile = "mosquitto.conf.template"
const DefaultAuthPlugin = "mosqauth.so"

// MMConfig with mosquitto manager plugin configuration
type MMConfig struct {
	// ACL file for authorization. Required.
	AclFile string `yaml:"aclFile"`

	// Server listening address. Required.
	Address string `yaml:"address"`

	// Path to CA certificate file. Required.
	CaCertFile string `yaml:"caCertFile"`

	// Unique client ID of service instance. Usually PluginID.
	// Defaults to 'PluginID'
	ClientID string `yaml:"clientID"`

	// Path to logging output folder. Required.
	LogFolder string `yaml:"logFolder"`

	// mqtt connect ports. Defaults from mqttclient
	MqttPortWS   int `yaml:"mqttPortWS"`
	MqttPortCert int `yaml:"mqttPortCert"`
	MqttPortUnpw int `yaml:"mqttPortUnpw"`

	// Path to the mosquitto authentication plugin. Required.
	MosqAuthPlugin string `yaml:"mosqAuthPlugin"`

	// filename of the generated mosquitto config file. Required.
	MosquittoConfFile string `yaml:"mosquittoConfFile"`

	// filename of mosquitto template. Required.
	MosquittoTemplateFile string `yaml:"mosquittoTemplateFile"`

	// Path to server certificate file. Required.
	ServerCertFile string `yaml:"serverCertFile"`

	// Path to server key file. Required.
	ServerKeyFile string `yaml:"serverKeyFile"`
}

// MosquittoManager manages configuration and launching of the mosquitto broker
type MosquittoManager struct {
	// Service configuration
	Config MMConfig
	// MQTT client to communicate with the Hub
	mqttClient *mqttclient.MqttClient // for communication with the Hub
	// Command that runs the Mosquitto broker
	mosquittoCmd *exec.Cmd
	// Flag, service is running
	isRunning chan bool
}

// Start the manager.
//
// Preconditions:
//  * Certificates must exist in the certificate folder from HubConfig
//  * ACL store must exist in the config folder
//
// This:
//  1. Generate a new mosquitto configuration file at ${configFolder}/mosquitto.conf
//  2. Listen for CLI commands to manage users and roles
//  3. Publish this service TD (if enabled) to make the mqtt broker discoverable
//
// Returns error if no mosquitto configuration is found
func (mm *MosquittoManager) Start() error {
	logrus.Warningf("Start")

	// make sure config is set
	if mm.Config.AclFile == "" ||
		mm.Config.Address == "" ||
		mm.Config.CaCertFile == "" ||
		mm.Config.LogFolder == "" ||
		mm.Config.MosqAuthPlugin == "" ||
		mm.Config.MosquittoConfFile == "" ||
		mm.Config.MosquittoTemplateFile == "" ||
		mm.Config.ServerCertFile == "" ||
		mm.Config.ServerKeyFile == "" {
		logrus.Fatalf("Mosquitto Manager config is missing fields.")
	}

	templateFilename := mm.Config.MosquittoTemplateFile
	err := ConfigureMosquitto(&mm.Config, templateFilename, mm.Config.MosquittoConfFile)
	if err != nil {
		return err
	}
	mm.mosquittoCmd, err = LaunchMosquitto(mm.Config.MosquittoConfFile, mm.isRunning)
	if err != nil {
		logrus.Errorf("Mosquitto failed to start: %s", err)
		return err
	}

	// This manager communicates with the Hub using the message bus
	// mm.hubClient = hubclient.NewMqttHubPluginClient(mm.Config.ClientID, mm.hubConfig)
	// err = mm.hubClient.Start()
	// if err != nil {
	// 	logrus.Errorf("MosquittoManager.Start MQTT client failed (for hub comm)")
	// 	mm.mosquittoCmd.Process.Kill()
	// 	mm.mosquittoCmd = nil
	// 	return err
	// }
	logrus.Infof("MosquittoManager.Start success")
	// Listen for provisioning requests
	// topic := MakeProvisionTopic('+', ProvisionRequest)
	// mm.hubClient.SubscribeProvisioning(topic, HandleProvisionRequest)
	//
	// mm.PublishServiceTD()

	return nil
}

// Stop the manager.
// If Mosquitto was started by the manager it will be stopped.
func (mm *MosquittoManager) Stop() {
	if mm.mosquittoCmd != nil {
		// err := mm.mosquittoCmd.Process.Kill()
		err := mm.mosquittoCmd.Process.Signal(os.Kill)
		if err != nil {
			logrus.Errorf("MosquittoManager.Stop. Kill mosquitto error: %s", err)
		}
		// FIXME: use channel to wait for completion
		<-mm.isRunning
		logrus.Warningf("MosquittoManager.Stop. Mosquitto ended")
		// _, err = mm.mosquittoCmd.Process.Wait()
		// if err != nil {
		// 	logrus.Infof("MosquittoManager.Stop. Wait mosquitto error: %s", err)
		// }
	}
	if mm.mqttClient != nil {
		mm.mqttClient.Disconnect()
	}
}

// NewMosquittoManager creates the mosquitto manager plugin
//  svcConfig contains the mosquitto manager configuration settings
func NewMosquittoManager(svcConfig MMConfig) *MosquittoManager {
	// set defaults
	if svcConfig.ClientID == "" {
		svcConfig.ClientID = PluginID
	}
	if svcConfig.MqttPortCert == 0 {
		svcConfig.MqttPortCert = config.DefaultMqttPortCert
	}
	if svcConfig.MqttPortUnpw == 0 {
		svcConfig.MqttPortUnpw = config.DefaultMqttPortUnpw
	}
	if svcConfig.MqttPortWS == 0 {
		svcConfig.MqttPortWS = config.DefaultMqttPortWS
	}
	mm := &MosquittoManager{
		Config:    svcConfig,
		isRunning: make(chan bool, 1),
	}
	return mm
}
