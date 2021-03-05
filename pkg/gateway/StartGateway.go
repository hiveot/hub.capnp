package gateway

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/config"
	"github.com/wostzone/gateway/pkg/messaging"
	"github.com/wostzone/gateway/pkg/smbserver"
)

// the internal message bus server, if running
var srv *smbserver.ServeSmbus

// StartGateway reads the gateway configuration, starts the internal message bus server, if configured,
// and launches the plugins.
// If the configuration is invalid then start is aborted
// The plugins receive the same commandline arguments as the gateway
//  homeFolder is the folder containing the config subfolder with the gateway.yaml configuration
//  startPlugins set to false to only start the gateway with message bus server if configured
func StartGateway(homeFolder string, startPlugins bool) error {
	var err error
	config, err := config.SetupConfig(homeFolder, "", nil)
	if err != nil {
		return err
	}
	if config.Messenger.Protocol != messaging.ConnectionProtocolMQTT {
		logrus.Warningf("StartGateway: Starting the internal message bus server")
		srv, err = smbserver.StartSmbServer(config)
	}

	// launch plugins
	logrus.Warningf("StartGateway: Starting %d gateway plugins on %s. UseTLS=%t",
		len(config.Plugins), config.Messenger.HostPort, config.Messenger.CertFolder != "")

	if startPlugins {
		args := os.Args[1:] // pass the gateways args to the plugin
		StartPlugins(config.PluginFolder, config.Plugins, args)
	}
	logrus.Warningf("StartGateway: Gateway started successfully!")

	return nil
}

// StopGateway stops a running gateway and its plugins
// TODO implements
func StopGateway() {
	logrus.Warningf("StopGateway: Received Signal, stopping gateway and its plugins")
	if srv != nil {
		srv.Stop()
	}
	logrus.Warningf("StopGateway: Unable to stop gateway plugins. Someone hasn't implemented this yet...")
}
