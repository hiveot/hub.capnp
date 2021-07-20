package mosquittomgr

import (
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

const PluginID = "mosquittomgr"

const DefaultConfFile = "mosquitto.conf"
const DefaultTemplateFile = "mosquitto.conf.template"

// Mosquitto manager configuration
type PluginConfig struct {
	ClientID          string `yaml:"clientID"`          // custom unique client ID of service instance
	MosquittoConf     string `yaml:"mosquittoConf"`     // generated mosquitto config file. Default is /etc/mosquitto/conf.d/wost.conf
	MosquittoTemplate string `yaml:"mosquittoTemplate"` // filename of mosquitto template
}

// Manager for mosquitto configuration
type MosquittoManager struct {
	Config       PluginConfig
	hubConfig    *hubconfig.HubConfig
	hubClient    *hubclient.MqttHubClient
	mosquittoCmd *exec.Cmd
}

// Start the manager.
// Installation preconditions:
//   Certificates must exist in the WoST certificate folder
//   Softlinks must have been created from mosquitto to wosthome/config and wosthome/certs
// This:
// 1. Generate a new mosquitto WoST configuration file in $wosthome/config/wost.conf
// 3. Generate Mosquitto ACL templates if they do not exist
// 4. Listen for CLI commands to manage users and roles
// 5. Publish this service TD (if enabled) to make the mqtt broker discoverable
//
//  hubConfig contains the WoST hub configuration with connection address and various folders
// Returns error if no mosquitto configuration is found
func (mm *MosquittoManager) Start(hubConfig *hubconfig.HubConfig) error {
	mm.hubConfig = hubConfig

	templateFilename := mm.Config.MosquittoTemplate
	configFile, err := ConfigureMosquitto(mm.hubConfig, templateFilename, mm.Config.MosquittoConf)
	if err != nil {
		return err
	}
	mm.mosquittoCmd, err = LaunchMosquitto(configFile)
	if err != nil {
		logrus.Fatalf("Mosquitto failed to start: %s", err)
		return err
	}

	mm.hubClient = hubclient.NewMqttHubPluginClient(mm.Config.ClientID, mm.hubConfig)
	err = mm.hubClient.Start()
	if err != nil {
		return err
	}
	logrus.Infof("MosquittoManager started successfully")
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
		mm.mosquittoCmd.Process.Kill()
	}
	if mm.hubClient != nil {
		mm.hubClient.Stop()
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
	}
	return mm
}
