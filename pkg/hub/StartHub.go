package hub

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/pkg/config"
	"github.com/wostzone/hub/pkg/messaging"
	"github.com/wostzone/hub/pkg/smbserver"
)

// the internal message bus server, if running
var srv *smbserver.ServeSmbus

// StartHub reads the hub configuration, starts the internal message bus server, if configured,
// and launches the plugins.
// If the configuration is invalid then start is aborted
// The plugins receive the same commandline arguments as the hub
//  homeFolder is the folder containing the config subfolder with the hub.yaml configuration
//  startPlugins set to false to only start the hub with message bus server if configured
func StartHub(homeFolder string, startPlugins bool) error {
	var err error
	config, err := config.SetupConfig(homeFolder, "", nil)
	if err != nil {
		return err
	}
	if config.Messenger.Protocol != messaging.ConnectionProtocolMQTT {
		logrus.Warningf("Starting the internal message bus server")
		srv, err = smbserver.StartSmbServer(config)
	}

	if !startPlugins || config.PluginFolder == "" {
		logrus.Infof("Not starting plugins")
	} else {
		// launch plugins
		logrus.Warningf("Starting %d hub plugins on %s. UseTLS=%t",
			len(config.Plugins), config.Messenger.HostPort, config.Messenger.CertFolder != "")

		args := os.Args[1:] // pass the hubs args to the plugin
		StartPlugins(config.PluginFolder, config.Plugins, args)
	}

	logrus.Warningf("Hub started successfully!")

	return nil
}

// StopHub stops a running hub and its plugins
// TODO implements
func StopHub() {
	logrus.Warningf("Received Signal, stopping hub and its plugins")
	if srv != nil {
		srv.Stop()
	}
	logrus.Warningf("Unable to stop hub plugins. Someone hasn't implemented this yet...")
}
