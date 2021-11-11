// Package config with the global hub configuration struct and methods
package config

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/certs"
)

// DefaultHubConfigName with the configuration file name of the hub
const DefaultHubConfigName = "hub.yaml"

// DefaultBinFolder is the location of application binaries wrt installation folder
const DefaultBinFolder = "./bin"

// DefaultCertsFolder with the location of certificates
const DefaultCertsFolder = "./certs"

// DefaultConfigFolder is the location of config files wrt installation folder
const DefaultConfigFolder = "./config"

// DefaultLogsFolder is the location of log files wrt wrt installation folder
const DefaultLogsFolder = "./logs"

// auth
// const (
// 	DefaultAclFile  = "hub.acl"
// 	DefaultUnpwFile = "hub.passwd"
// )

// Default ports for connecting to the MQTT server
const (
	DefaultMqttPortUnpw = 8883
	DefaultMqttPortCert = 8884
	DefaultMqttPortWS   = 8885
)
const DefaultMqttTimeout = 3

// Default certificate and private key file names
const (
	DefaultCaCertFile     = "caCert.pem"
	DefaultCaKeyFile      = "caKey.pem"
	DefaultPluginCertFile = "pluginCert.pem"
	DefaultPluginKeyFile  = "pluginKey.pem"
	DefaultServerCertFile = "serverCert.pem"
	DefaultServerKeyFile  = "serverKey.pem"
	DefaultAdminCertFile  = "adminCert.pem"
	DefaultAdminKeyFile   = "adminKey.pem"
)

// DefaultThingZone is the zone the hub's things are published in.
const DefaultThingZone = "local"

// HubConfig contains the global configuration for using the Hub by its clients.
//
// Intended for use by:
//  1. Hub plugins that needs to know the location of files, certificates and service address and ports
//  2. Remote devices or services that uses a local copy of the hub config for manual configuration of
//     certificates MQTT server address and ports.
//
type HubConfig struct {

	// MQTT server address. The default "" is localhost
	MqttAddress string `yaml:"mqttAddress,omitempty"`
	// MQTT TLS port for certificate based authentication. Default is DefaultMqttPortCert
	MqttPortCert int `yaml:"mqttPortCert,omitempty"`
	// MQTT TLS port for login/password authentication. Default is DefaultMqttPortUnpw
	MqttPortUnpw int `yaml:"mqttPortUnpw,omitempty"`
	// Websocket TLS port for login/password authentication. Default is DefaultMqttPortWS
	MqttPortWS int `yaml:"mqttPortWS,omitempty"`
	// plugin mqtt connection timeout in seconds. 0 for indefinite. Default is DefaultMqttTimeout (3 sec)
	MqttTimeout int `yaml:"mqttTimeout,omitempty"`

	// auth
	// AclStorePath  string `yaml:"aclStore"`  // path to the ACL store
	// UnpwStorePath string `yaml:"unpwStore"` // path to the uername/password store

	// Zone that published Things belong to. Default is 'local'
	// Zones are useful for separating devices from Hubs on large networks. Normally 'local' is sufficient.
	//
	// When Things are bridged, the bridge can be configured to replace the zone by that of the bridge.
	// This is intended for access control to Things from a different zone.
	Zone string `yaml:"zone"`

	// Files and Folders
	Loglevel    string `yaml:"logLevel"`    // debug, info, warning, error. Default is warning
	LogsFolder  string `yaml:"logsFolder"`  // location of Wost log files
	LogFile     string `yaml:"logFile"`     // log filename is pluginID.log
	AppFolder   string `yaml:"appFolder"`   // Folder containing the application installation
	BinFolder   string `yaml:"binFolder"`   // Folder containing plugin binaries, default is {appFolder}/bin
	CertsFolder string `yaml:"certsFolder"` // Folder containing certificates, default is {appFolder}/certs
	// ConfigFolder the location of additional configuration files. Default is {appFolder}/config
	ConfigFolder string `yaml:"configFolder"`

	// path to CA certificate in PEM format. Default is certs/caCert.pem
	CaCertPath string `yaml:"caCertPath"`
	// path to client x509 certificate in PEM format. Default is certs/{clientID}Cert.pem
	ClientCertPath string `yaml:"clientCertPath"`
	// path to client private key in PEM format. Default is certs/{clientID}Key.pem
	ClientKeyPath string `yaml:"clientKeyPath"`
	// path to plugin x509 certificate in PEM format. Default is certs/PluginCert.pem
	PluginCertPath string `yaml:"pluginCertPath"`
	// path to plugin private key in PEM format. Default is certs/PluginKey.pem
	PluginKeyPath string `yaml:"pluginKeyPath"`

	// CaCert contains the loaded CA certificate needed for establishing trusted connections to the
	// MQTT message bus and other services. Loading takes place in LoadHubConfig()
	CaCert *x509.Certificate

	// ClientCert contains the loaded TLS client certificate and key if available.
	// Loading takes place in LoadHubConfig()
	// * For plugins this is the plugin certificate and private key
	// * For servers this is the server certificate and private key
	// * For devices this is the provisioned device certificate and private key
	ClientCert *tls.Certificate

	// PluginCert contains the TLS client certificate for use by plugins
	// Intended for use by plugin clients. This is nil of the plugin certificate is not available or accessible
	// Loading takes place in LoadHubConfig()
	PluginCert *tls.Certificate
}

// CreateDefaultHubConfig with default values
// appFolder is the hub installation folder and home to plugins, logs and configuration folders.
// Use "" for default: parent of application binary
// When relative path is given, it is relative to the application binary
//
// See also LoadHubConfig to load the actual configuration including certificates.
func CreateDefaultHubConfig(appFolder string) *HubConfig {
	appBin, _ := os.Executable()
	binFolder := path.Dir(appBin)
	if appFolder == "" {
		appFolder = path.Dir(binFolder)
	} else if !path.IsAbs(appFolder) {
		// turn relative home folder in absolute path
		appFolder = path.Join(binFolder, appFolder)
	}
	logrus.Infof("AppBin is: %s; Home is: %s", appBin, appFolder)
	config := &HubConfig{
		AppFolder:    appFolder,
		BinFolder:    path.Join(appFolder, DefaultBinFolder),
		CertsFolder:  path.Join(appFolder, DefaultCertsFolder),
		ConfigFolder: path.Join(appFolder, DefaultConfigFolder),
		LogsFolder:   path.Join(appFolder, DefaultLogsFolder),
		Loglevel:     "warning",

		MqttAddress:  "127.0.0.1",
		MqttPortCert: DefaultMqttPortCert,
		MqttPortUnpw: DefaultMqttPortUnpw,
		MqttPortWS:   DefaultMqttPortWS,
		// Plugins:      make([]string, 0),
		Zone: "local",
	}
	// config.Messenger.CertsFolder = path.Join(homeFolder, "certs")
	// config.AclStorePath = path.Join(config.ConfigFolder, DefaultAclFile)
	// config.UnpwStorePath = path.Join(config.ConfigFolder, DefaultUnpwFile)
	return config
}

// LoadHubConfig loads and validates the global hub configuration; loads certificates and sets logging.
// Intended to be used after CreateDefaultHubConfig()
//
// Each client loads the global hub.yaml configuration.
// The following variables can be used in this file:
//    {clientID}  is the device or plugin instance ID. Used for logfile and client cert
//    {appFolder} is the default application folder (parent of application binary)
//    {certsFolder} is the default certificate folder
//    {configFolder} is the default configuration folder
//    {logsFolder} is the default logging folder
//
//  configFile is optional. By default this loads hub.yaml in the default config folder of hubConfig.
// Returns the hub configuration and error code in case of error
func LoadHubConfig(configFile string, clientID string, hubConfig *HubConfig) error {
	substituteMap := make(map[string]string)
	substituteMap["{clientID}"] = clientID
	substituteMap["{appFolder}"] = hubConfig.AppFolder
	substituteMap["{configFolder}"] = hubConfig.ConfigFolder
	substituteMap["{logsFolder}"] = hubConfig.LogsFolder
	substituteMap["{certsFolder}"] = hubConfig.CertsFolder

	if configFile == "" {
		configFile = path.Join(hubConfig.ConfigFolder, DefaultHubConfigName)
	}
	logrus.Infof("Using %s as hub config file", configFile)
	err := LoadYamlConfig(configFile, hubConfig, substituteMap)
	if err != nil {
		return err
	}

	// make sure files and folders have an absolute path
	if !path.IsAbs(hubConfig.CertsFolder) {
		hubConfig.CertsFolder = path.Join(hubConfig.AppFolder, hubConfig.CertsFolder)
	}

	if !path.IsAbs(hubConfig.LogsFolder) {
		hubConfig.LogsFolder = path.Join(hubConfig.AppFolder, hubConfig.LogsFolder)
	}

	if hubConfig.LogFile == "" {
		hubConfig.LogFile = path.Join(hubConfig.LogsFolder, clientID+".log")
	} else if !path.IsAbs(hubConfig.LogFile) {
		hubConfig.LogFile = path.Join(hubConfig.LogsFolder, hubConfig.LogFile)
	}
	SetLogging(hubConfig.Loglevel, hubConfig.LogFile)

	if !path.IsAbs(hubConfig.ConfigFolder) {
		hubConfig.ConfigFolder = path.Join(hubConfig.AppFolder, hubConfig.ConfigFolder)
	}

	// CA certificate for use by everyone
	if hubConfig.CaCertPath == "" {
		hubConfig.CaCertPath = path.Join(hubConfig.CertsFolder, DefaultCaCertFile)
	} else if !path.IsAbs(hubConfig.CaCertPath) {
		hubConfig.CaCertPath = path.Join(hubConfig.CertsFolder, hubConfig.CaCertPath)
	}

	// Plugin client certificate for use by plugin clients
	if hubConfig.PluginCertPath == "" {
		hubConfig.PluginCertPath = path.Join(hubConfig.CertsFolder, DefaultPluginCertFile)
	} else if !path.IsAbs(hubConfig.PluginCertPath) {
		hubConfig.PluginCertPath = path.Join(hubConfig.CertsFolder, hubConfig.PluginCertPath)
	}
	if hubConfig.PluginKeyPath == "" {
		hubConfig.PluginKeyPath = path.Join(hubConfig.CertsFolder, DefaultPluginKeyFile)
	} else if !path.IsAbs(hubConfig.PluginKeyPath) {
		hubConfig.PluginKeyPath = path.Join(hubConfig.CertsFolder, hubConfig.PluginKeyPath)
	}

	// Client certificate for use by clients with their own certificate, eg iot devices
	if hubConfig.ClientCertPath == "" {
		hubConfig.ClientCertPath = path.Join(hubConfig.CertsFolder, clientID+"Cert.pem")
	} else if !path.IsAbs(hubConfig.ClientCertPath) {
		hubConfig.ClientCertPath = path.Join(hubConfig.CertsFolder, hubConfig.ClientCertPath)
	}

	if hubConfig.ClientKeyPath == "" {
		hubConfig.ClientKeyPath = path.Join(hubConfig.CertsFolder, clientID+"Key.pem")
	} else if !path.IsAbs(hubConfig.ClientCertPath) {
		hubConfig.ClientKeyPath = path.Join(hubConfig.CertsFolder, hubConfig.ClientKeyPath)
	}

	// Certificate are optional as they might not yet exist
	hubConfig.CaCert, err = certs.LoadX509CertFromPEM(hubConfig.CaCertPath)
	if err != nil {
		logrus.Warningf("LoadHubConfig: Unable to load the CA Certificate: %s. This is not good but continuing for now.", err)
	}
	// optional client certificate, if available
	hubConfig.ClientCert, err = certs.LoadTLSCertFromPEM(hubConfig.ClientCertPath, hubConfig.ClientKeyPath)
	if err != nil {
		logrus.Warningf("LoadHubConfig: Unable to load the Client Certificate: %s. This is only needed when not a plugin so continuing for now.", err)
	}
	// optional plugin certificate, if available
	hubConfig.PluginCert, err = certs.LoadTLSCertFromPEM(hubConfig.PluginCertPath, hubConfig.PluginKeyPath)
	if err != nil {
		logrus.Warningf("LoadHubConfig: Unable to load the Plugin Certificate: %s. This is only needed for plugins so continuing for now.", err)
	}

	// validate the result
	err = ValidateHubConfig(hubConfig)
	return err
}

// ValidateHubConfig checks if values in the hub configuration are correct
// Returns an error if the config is invalid
func ValidateHubConfig(config *HubConfig) error {
	if _, err := os.Stat(config.AppFolder); os.IsNotExist(err) {
		logrus.Errorf("Home folder '%s' not found\n", config.AppFolder)
		return err
	}
	if _, err := os.Stat(config.ConfigFolder); os.IsNotExist(err) {
		logrus.Errorf("Configuration folder '%s' not found\n", config.ConfigFolder)
		return err
	}

	if _, err := os.Stat(config.LogsFolder); os.IsNotExist(err) {
		logrus.Errorf("Logging folder '%s' not found\n", config.LogsFolder)
		return err
	}

	if _, err := os.Stat(config.CertsFolder); os.IsNotExist(err) {
		logrus.Errorf("TLS certificate folder '%s' not found\n", config.CertsFolder)
		return err
	}

	return nil
}
