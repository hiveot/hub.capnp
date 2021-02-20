package lib

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/messaging"
	"github.com/wostzone/gateway/pkg/messaging/smbus"
	"gopkg.in/yaml.v2"
)

// ConfigArgs configuration commandline arguments
type ConfigArgs struct {
	name         string
	defaultValue string
	description  string
}

// GatewayConfig with gateway configuration parameters
type GatewayConfig struct {
	Logging struct {
		Loglevel   string `yaml:"loglevel"`   // debug, info, warning, error. Default is warning
		LogsFolder string `yaml:"logsFolder"` // location of logfiles
	} `yaml:"logging"`

	// Messenger configuration of gateway plugin messaging
	Messenger struct {
		CertsFolder string `yaml:"certsFolder"` // location of gateway and client certificate
		HostPort    string `yaml:"hostname"`    // hostname:port or ip:port to listen on of message bus
		Protocol    string `yaml:"protocol"`    // internal, MQTT, default internal
		UseTLS      bool   `yaml:"useTLS"`      // use TLS for client/server messaging
	} `yaml:"messenger"`

	// AppFolder    string   `yaml:"app"`          // application root folder
	ConfigFolder string   `yaml:"configFolder"` // location of plugin configuration files
	PluginFolder string   `yaml:"pluginFolder"` // location of plugin binaries
	Plugins      []string `yaml:"plugins"`      // names of plugins to start
	// internal
}

// CreateGatewayConfig with default values
// baseFolder is the base of the configuration. Use "" for default: parent of application
//
func CreateGatewayConfig(baseFolder string) *GatewayConfig {
	appFolder := baseFolder
	if appFolder == "" {
		appBin, _ := os.Executable()
		binFolder := path.Dir(appBin)
		appFolder = path.Dir(binFolder)

		// for running within the project use the test folder as application root folder
		if path.Base(binFolder) != "bin" {
			appFolder = path.Join(appFolder, "test")
		}
	}
	config := &GatewayConfig{
		ConfigFolder: path.Join(appFolder, "config"),
		Plugins:      make([]string, 0),
		PluginFolder: path.Join(appFolder, "bin"),
	}
	config.Messenger.CertsFolder = path.Join(appFolder, "certs")
	config.Messenger.HostPort = smbus.DefaultSmbusHost                    //"localhost:9678"
	config.Messenger.Protocol = string(messaging.ConnectionProtocolSmbus) // internal
	config.Messenger.UseTLS = true
	config.Logging.Loglevel = "warning"
	config.Logging.LogsFolder = path.Join(appFolder, "logs")
	return config
}

// LoadConfig loads the configuration from file into the given config
// Returns nil if successful
func LoadConfig(configFile string, config interface{}) error {
	var err error
	var rawConfig []byte
	rawConfig, err = ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Errorf("LoadConfig: Error loading config from file '%s': %s", configFile, err)
		return err
	}
	logrus.Infof("LoadConfig: Loaded config from file '%s'", configFile)
	err = yaml.Unmarshal(rawConfig, config)
	if err != nil {
		logrus.Errorf("LoadConfig: Error parsing config file '%s': %s", configFile, err)
		return err
	}
	return nil
}

// ValidateConfig checks if values in the gatewy configuration are correct
// Returns an error if the config is invalid
func ValidateConfig(config *GatewayConfig) error {
	// validate config file
	if _, err := os.Stat(config.ConfigFolder); os.IsNotExist(err) {
		logrus.Errorf("Configuration folder '%s' not found\n", config.ConfigFolder)
		return err
	}
	if _, err := os.Stat(config.Logging.LogsFolder); os.IsNotExist(err) {
		logrus.Errorf("Logging folder '%s' not found\n", config.Logging.LogsFolder)
		return err
	}
	if _, err := os.Stat(config.Messenger.CertsFolder); os.IsNotExist(err) && config.Messenger.UseTLS {
		logrus.Errorf("TLS certificate folder '%s' not found\n", config.Messenger.CertsFolder)
		return err
	}
	if _, err := os.Stat(config.PluginFolder); os.IsNotExist(err) {
		logrus.Errorf("Plugins folder '%s' not found\n", config.PluginFolder)
		return err
	}
	return nil
}

// // Set default configuration and load optional configuration file
// func loadConfig(configFile string) *Config {
// 	gwbin, _ := os.Executable()
// 	binFolder := path.Dir(gwbin)
// 	appFolder := path.Dir(binFolder)
// 	// for running within the project use the test folder as application root folder
// 	if path.Base(binFolder) != "bin" {
// 		appFolder = path.Join(appFolder, "test")
// 	}
// 	config := &Config{
// 		Channels:   []string{messaging.TDChannelID, messaging.ActionChannelID, messaging.EventsChannelID},
// 		LogsFolder: path.Join(appFolder, "logs"),
// 	}

// 	// configFile := path.Join(config.ConfigFolder, "gateway.yaml")
// 	rawConfig, err := ioutil.ReadFile(configFile)
// 	if err == nil {
// 		logrus.Infof("Loading configuration from: %s", configFile)
// 		err = yaml.Unmarshal(rawConfig, config)
// 		if err != nil {
// 			logrus.Errorf("Failed parsing configuration file %s: %s", configFile, err)
// 		}
// 	}
// 	return config
// }
