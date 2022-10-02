package internal_old

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

// LauncherConfig holds the launcher configuration of services
type LauncherConfig struct {
	// Core services
	Services map[string]ServiceConfig `yaml:"services"`
	// Gateway services
	Gateway map[string]ServiceConfig `yaml:"gateway"`
	// Bindings running in the sidecar
	Bindings map[string]BindingConfig `yaml:"bindings"`
}

type ServiceConfig struct {
	// Application settings
	ServiceBin      string `yaml:"svc-bin"`      // path to service binary
	ServiceID       string `yaml:"svc-id"`       // unique name of the service
	ServicePort     int    `yaml:"svc-port"`     // invoke service on this port
	ServiceProtocol string `yaml:"svc-protocol"` // invoke service using 'http' or 'grpc' protocol
	ServiceSocket   string `yaml:"svc-socket"`   // invoke service using unix domain sockets
	UseSSL          bool   `yaml:"svc-ssl"`      // invoke service using http over ssl
	AutoStart       bool   `yaml:"autoStart"`    // start service when launching the hub

	// Dapr settings
	DaprComponents string `yaml:"dapr-components"` // dapr components folder
	DaprConfig     string `yaml:"dapr-config"`     // dapr config file
	DaprHttpPort   int    `yaml:"dapr-http-port"`  // dapr http listening port
	DaprGrpcPort   int    `yaml:"dapr-grpc-port"`  // dapr grpc listening port
	LogLevel       string `yaml:"loglevel"`        // dapr logging: debug, info, warn, error, fatal or panic
}

// BindingConfig Bindings run in the sidecar. Each service can have its own bindings configuration
type BindingConfig struct {
	Name string `yaml:"name"`
}

// Load the launcher configuration from yaml file
func (config *LauncherConfig) Load(path string) error {

	rawConfig, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Infof("LoadYamlConfig: Error loading config file: %s: %s", path, err)
		return err
	}
	logrus.Infof("Loaded config file '%s'", path)

	err = yaml.Unmarshal(rawConfig, config)
	if err != nil {
		logrus.Errorf("LoadYamlConfig: Error parsing config file '%s': %s", path, err)
		return err
	}
	return nil
}
