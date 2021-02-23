package gateway

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
)

// StartGateway reads the gateway configuration and starts the gateway plugins
// Start is aborted if the configuration is invalid
func StartGateway(appFolder string) error {
	config, err := lib.SetupConfig(appFolder, "", nil)
	if err != nil {
		return err
	}

	// launch plugins
	logrus.Warningf("Starting %d gateway plugins on %s. UseTLS=%t",
		len(config.Plugins), config.Messenger.HostPort, config.Messenger.UseTLS)
	args := os.Args[1:] // pass the gateways args to the plugin

	lib.StartPlugins(config.PluginFolder, config.Plugins, args)

	logrus.Warningf("Gateway is not running")

	return nil
}

// StopGateway stops a running gateway and its plugins
// TODO implements
func StopGateway() {
	logrus.Warningf("Received Signal, stopping gateway and its plugins")

	logrus.Warningf("Unable to stop gateway plugins. Someone hasn't implemented this yet...")
}
