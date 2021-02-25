package lib

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/messaging"
	"github.com/wostzone/gateway/pkg/messaging/smbus"
	"gopkg.in/yaml.v2"
)

// GatewayConfigName the configuration file name of the gateway
const GatewayConfigName = "gateway.yaml"

// GatewayLogFile the file name of the gateway logging
const GatewayLogFile = "gateway.log"

// ConfigArgs configuration commandline arguments
type ConfigArgs struct {
	name         string
	defaultValue string
	description  string
}

// GatewayConfig with gateway configuration parameters
type GatewayConfig struct {
	Logging struct {
		Loglevel string `yaml:"logLevel"` // debug, info, warning, error. Default is warning
		LogFile  string `yaml:"logFile"`  // gateway logging to file
	} `yaml:"logging"`

	// Messenger configuration of gateway plugin messaging
	Messenger struct {
		CertsFolder string `yaml:"certsFolder"` // location of gateway and client certificate. Default is ./certs
		HostPort    string `yaml:"hostname"`    // hostname:port or ip:port to listen on of message bus
		Protocol    string `yaml:"protocol"`    // internal, MQTT, default internal
		UseTLS      bool   `yaml:"useTLS"`      // use TLS for client/server messaging
	} `yaml:"messenger"`

	Home         string   `yaml:"home"`         // application home directory. Default is parent of executable.
	ConfigFolder string   `yaml:"configFolder"` // location of configuration files. Default is ./config
	PluginFolder string   `yaml:"pluginFolder"` // location of plugin binaries. Default is ./bin
	Plugins      []string `yaml:"plugins"`      // names of plugins to start
	// internal
}

// CreateDefaultGatewayConfig with default values
// homeFolder is the home of the application, log and configuration folders.
// Use "" for default: parent of application binary
// When relative path is given, it is relative to the application binary
func CreateDefaultGatewayConfig(homeFolder string) *GatewayConfig {
	appBin, _ := os.Executable()
	binFolder := path.Dir(appBin)
	if homeFolder == "" {
		homeFolder = path.Dir(binFolder)

		// for running within the project tests use the test folder as application root folder
		// if path.Base(binFolder) != "bin" {
		// 	// appFolder = path.Join(appFolder, "../test")
		// 	cwd, _ := os.Getwd()
		// 	appFolder = path.Join(cwd, "../../test")
		// }
		// logrus.Infof("appBin: %s. CWD=%s", appBin, cwd)
	} else if !path.IsAbs(homeFolder) {
		// turn relative home folder in absolute path
		// cwd, _ := os.Getwd()
		// homeFolder = path.Join(cwd, homeFolder)
		homeFolder = path.Join(binFolder, homeFolder)
	}
	logrus.Infof("CreateDefaultGatewayConfig: appBin is: %s; Home is: %s", appBin, homeFolder)
	config := &GatewayConfig{
		// ConfigFolder: path.Join(homeFolder, "config"),
		Home:         homeFolder,
		ConfigFolder: path.Join(homeFolder, "config"),
		Plugins:      make([]string, 0),
		// PluginFolder: path.Join(homeFolder, "bin"),
		PluginFolder: path.Join(homeFolder, "./bin"),
	}
	// config.Messenger.CertsFolder = path.Join(homeFolder, "certs")
	config.Messenger.CertsFolder = path.Join(homeFolder, "./certs")
	config.Messenger.HostPort = smbus.DefaultSmbusHost                    //"localhost:9678"
	config.Messenger.Protocol = string(messaging.ConnectionProtocolSmbus) // internal
	config.Messenger.UseTLS = true
	config.Logging.Loglevel = "warning"
	// config.Logging.LogFile = path.Join(homeFolder, "logs/"+GatewayLogFile)
	config.Logging.LogFile = path.Join(homeFolder, "./logs/"+GatewayLogFile)
	return config
}

// LoadConfig loads the configuration from file into the given config
// Returns nil if successful
func LoadConfig(configFile string, config interface{}) error {
	var err error
	var rawConfig []byte
	rawConfig, err = ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Warningf("LoadConfig: Unable to load config file: %s", err)
		return err
	}
	logrus.Infof("LoadConfig: Loaded config file '%s'", configFile)

	err = yaml.Unmarshal(rawConfig, config)
	if err != nil {
		logrus.Errorf("LoadConfig: Error parsing config file '%s': %s", configFile, err)
		return err
	}
	return nil
}

// ValidateConfig checks if values in the gateway configuration are correct
// Returns an error if the config is invalid
func ValidateConfig(config *GatewayConfig) error {
	if _, err := os.Stat(config.Home); os.IsNotExist(err) {
		logrus.Errorf("ValidateConfig: Home folder '%s' not found\n", config.Home)
		return err
	}
	if _, err := os.Stat(config.ConfigFolder); os.IsNotExist(err) {
		logrus.Errorf("ValidateConfig: Configuration folder '%s' not found\n", config.ConfigFolder)
		return err
	}

	loggingFolder := path.Dir(config.Logging.LogFile)
	if _, err := os.Stat(loggingFolder); os.IsNotExist(err) {
		logrus.Errorf("ValidateConfig: Logging folder '%s' not found\n", loggingFolder)
		return err
	}

	if _, err := os.Stat(config.Messenger.CertsFolder); os.IsNotExist(err) && config.Messenger.UseTLS {
		logrus.Errorf("ValidateConfig: TLS certificate folder '%s' not found\n", config.Messenger.CertsFolder)
		return err
	}

	if _, err := os.Stat(config.PluginFolder); os.IsNotExist(err) {
		logrus.Errorf("ValidateConfig: Plugins folder '%s' not found\n", config.PluginFolder)
		return err
	}

	if config.Messenger.Protocol != "" &&
		config.Messenger.Protocol != string(messaging.ConnectionProtocolSmbus) &&
		config.Messenger.Protocol != string(messaging.ConnectionProtocolMQTT) {
		err := errors.New("Invalid messenger protocol " + string(config.Messenger.Protocol))
		logrus.Errorf("ValidateConfig: %s", err)
		return err
	}

	return nil
}
