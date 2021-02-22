package internal

import "github.com/wostzone/gateway/pkg/lib"

// PluginID is the ID of the plugin
const PluginID = "logger"

// Config is the GatewayConfig extended with logging channels
type Config struct {
	lib.GatewayConfig          // embedded
	Channels          []string `yaml:"channels"`
}

// Plugin is a gateway plugin for logging channels
type Plugin struct {
}

// Start the logging
func (lp *Plugin) Start() {

}

// Stop the logging
func (lp *Plugin) Stop() {

}

// NewLoggerPlugin creates a new logging plugin instance
func NewLoggerPlugin(config *Config) *Plugin {
	return &Plugin{}
}
