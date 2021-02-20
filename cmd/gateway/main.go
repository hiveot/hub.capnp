package main

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
)

// StartGateway reads configurationand starts the gateway plugins
// Start is aborted if the configuration is invalid
func StartGateway() {
	var configFile string

	config := lib.CreateGatewayConfig("")
	lib.SetupGatewayArgs(config)
	flag.Parse()

	// optionally load a different config
	configFile = path.Join(config.ConfigFolder, "gateway.yaml")

	err1 := lib.LoadConfig(configFile, config)
	err2 := lib.ValidateConfig(config)
	// Configuration must be valid to continue
	if err1 != nil || err2 != nil {
		os.Exit(1)
	}
	lib.SetLogging(config.Logging.Loglevel, path.Join(config.Logging.LogsFolder, "gateway.log"))

	// launch plugins
	logrus.Warningf("Starting %d gateway plugins on %s. UseTLS=%t",
		len(config.Plugins), config.Messenger.HostPort, config.Messenger.UseTLS)
	args := os.Args[1:] // pass the gateways args to the plugin
	lib.StartPlugins(config.PluginFolder, config.Plugins, args)
	// TODO: check for valid connection to configured server
}

func main() {
	StartGateway()
}
