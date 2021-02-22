package gateway

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
)

// StartGateway reads the gateway configuration and starts the gateway plugins
// Start is aborted if the configuration is invalid
func StartGateway() error {
	config, err := lib.SetupConfig("", nil)
	if err != nil {
		return err
	}
	logFileName := path.Join(config.Logging.LogsFolder, "gateway.log")
	lib.SetLogging(config.Logging.Loglevel, logFileName)

	// launch plugins
	logrus.Warningf("Starting %d gateway plugins on %s. UseTLS=%t",
		len(config.Plugins), config.Messenger.HostPort, config.Messenger.UseTLS)
	args := os.Args[1:] // pass the gateways args to the plugin
	lib.StartPlugins(config.PluginFolder, config.Plugins, args)
	// TODO: check for valid connection to configured server
	return nil
}
