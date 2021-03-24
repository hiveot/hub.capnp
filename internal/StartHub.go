package hub

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/pkg/certsetup"
	"github.com/wostzone/hubapi/pkg/hubconfig"
)

// StartHub reads the hub configuration, create the certificates if needed and
// launches the plugins. If the configuration is invalid then start is aborted
// The plugins receive the same commandline arguments as the hub
//  homeFolder is the folder containing the config subfolder with the hub.yaml configuration
//  startPlugins set to false to only start the hub with message bus server if configured
// Return nil or error if the hub configuration file or certificate are not found
func StartHub(homeFolder string, startPlugins bool) error {
	var err error
	var noPlugins bool
	flag.BoolVar(&noPlugins, "noplugins", !startPlugins, "Start the hub without plugins")
	hc, err := hubconfig.LoadCommandlineConfig(homeFolder, "", nil)
	if err != nil {
		return err
	}
	// Exit if certificates don't exist
	caCertFile := path.Join(hc.Messenger.CertsFolder, certsetup.CaCertFile)
	if _, err := os.Stat(caCertFile); os.IsNotExist(err) {
		logrus.Fatalf("CA Certificate file %s not found.", caCertFile)
		return err
	} else {
		logrus.Warningf("Using certificates from %s", hc.Messenger.CertsFolder)
	}

	// start the plugins unless disabled
	if noPlugins || hc.PluginFolder == "" {
		logrus.Warningf("Starting Hub without plugins")
	} else {
		// launch plugins
		logrus.Warningf("Starting %d plugins on %s:%d.",
			len(hc.Plugins), hc.Messenger.Address, hc.Messenger.Port)

		args := os.Args[1:] // pass the hubs args to the plugin
		StartPlugins(hc.PluginFolder, hc.Plugins, args)
	}

	logrus.Warningf("Hub started successfully!")

	return nil
}

// StopHub stops a running hub and its plugins
// TODO implements
func StopHub() {
	logrus.Warningf("Received Signal, stopping hub and its plugins")
	logrus.Warningf("Unable to stop hub plugins. Someone hasn't implemented this yet...")
}
