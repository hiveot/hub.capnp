package hub

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

const pluginID = "hub"

// StartHub reads the hub configuration and launches the plugins. If the configuration is invalid
// then start is aborted. The plugins receive the same commandline arguments as the hub.
// Before starting the Hub, the certificates must have been generated as part of setup.
// Use 'gencerts' to generate them in the {homeFolder}/certs folder.
//  homeFolder is the parent folder of the application binary and contains the config subfolder
//  startPlugins set to false to only start the hub with message bus server if configured
// Return nil or error if the hub configuration file or certificate are not found
func StartHub(homeFolder string, startPlugins bool) error {

	var err error
	var noPlugins bool
	// the noplugins commandline argument only applies to the hub
	flag.BoolVar(&noPlugins, "noplugins", !startPlugins, "Start the hub without plugins")
	hc, err := hubconfig.LoadCommandlineConfig(homeFolder, pluginID, nil)

	if err != nil {
		return err
	}
	pluginFolder := path.Join(hc.Home, "bin")
	fmt.Printf("Home=%s\nPluginFolder=%s\n", hc.Home, pluginFolder)

	// Create a CA if needed and update hub and plugin certs
	sanNames := []string{hc.MqttAddress}
	certsetup.CreateCertificateBundle(sanNames, hc.CertsFolder)

	// start the plugins unless disabled
	if noPlugins {
		logrus.Infof("Starting Hub without plugins")
	} else {
		// launch plugins
		logrus.Infof("Starting %d plugins on %s.", len(hc.Plugins), hc.MqttAddress)

		args := os.Args[1:] // pass the hubs args to the plugin
		StartPlugins(pluginFolder, hc.Plugins, args)
	}

	logrus.Warningf("Hub start successful!")

	return nil
}

// StopHub stops a running hub and its plugins
// TODO implements
func StopHub() {
	logrus.Warningf("Received Signal, stopping hub and its plugins")
	logrus.Warningf("Unable to stop hub plugins. Someone hasn't implemented this yet...")
}
