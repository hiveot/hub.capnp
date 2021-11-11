package internal

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/mqttclient"
)

const PluginID = "mosquittomgr"

const DefaultConfFile = "mosquitto.conf"
const DefaultTemplateFile = "mosquitto.conf.template"

// PluginConfig with mosquitto manager configuration
type PluginConfig struct {
	ClientID          string `yaml:"clientID"`          // custom unique client ID of service instance
	MosquittoConf     string `yaml:"mosquittoConf"`     // generated mosquitto config file. Default is /etc/mosquitto/conf.d/wost.conf
	MosquittoTemplate string `yaml:"mosquittoTemplate"` // filename of mosquitto template
}

// MosquittoManager is the manager for mosquitto configuration
type MosquittoManager struct {
	Config       PluginConfig
	hubConfig    *config.HubConfig
	hubClient    *mqttclient.MqttHubClient // for communication with the Hub
	mosquittoCmd *exec.Cmd
	isRunning    chan bool
}

// Start the manager.
//
// Preconditions:
//   Certificates must exist in the certificate folder
//   Password and ACL store must exist in the config folder
//
// This:
// 1. Generate a new mosquitto configuration file at $wosthome/config/mosquitto.conf
// 2. Listen for CLI commands to manage users and roles
// 3. Publish this service TD (if enabled) to make the mqtt broker discoverable
//
//  hubConfig contains the WoST hub configuration with connection address and various folders
// Returns error if no mosquitto configuration is found
func (mm *MosquittoManager) Start(hubConfig *config.HubConfig) error {
	mm.hubConfig = hubConfig
	logrus.Warningf("MosquittoManager.Start")

	templateFilename := mm.Config.MosquittoTemplate
	configFile, err := ConfigureMosquitto(mm.hubConfig, templateFilename, mm.Config.MosquittoConf)
	if err != nil {
		return err
	}
	mm.mosquittoCmd, err = LaunchMosquitto(configFile, mm.isRunning)
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
	if mm.hubClient != nil {
		mm.hubClient.Close()
	}
}

// Start the mosquitto manager plugin
func NewMosquittoManager() *MosquittoManager {
	mm := &MosquittoManager{
		Config: PluginConfig{
			ClientID:          PluginID,
			MosquittoConf:     DefaultConfFile,
			MosquittoTemplate: DefaultTemplateFile,
		},
		isRunning: make(chan bool, 1),
	}
	return mm
}
