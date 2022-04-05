package internal

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/certs/pkg/certsetup"
	"github.com/wostzone/hub/lib/client/pkg/config"
)

// This is the Hub launcher
const pluginID = "launcher"

// PluginConfig with list of plugins to launch
type PluginConfig struct {
	Plugins []string `yaml:"plugins"`
}

// StartHub reads the launcher configuration and launches the plugins. If the configuration is invalid
// then start is aborted. The plugins receive the same commandline arguments as the launcher.
// Before starting the Hub, the certificates must have been generated as part of setup.
// Use 'gencerts' to generate them in the {homeFolder}/certsclient folder.
//  homeFolder is the parent folder of the application binary and contains the config subfolder
//  startPlugins set to false to only start the launcher with message bus server if configured
// Return nil or error if the launcher configuration file or certificate are not found
func StartHub(homeFolder string, startPlugins bool) error {

	var err error
	var noPlugins bool
	var pluginConfig PluginConfig

	// the noplugins commandline argument only applies to the launcher
	flag.BoolVar(&noPlugins, "noplugins", !startPlugins, "Start the launcher without plugins")
	hc, err := config.LoadAllConfig(os.Args, homeFolder, pluginID, &pluginConfig)

	if err != nil {
		return err
	}
	pluginFolder := path.Join(hc.AppFolder, "bin")
	fmt.Printf("Home=%s\nPluginFolder=%s\n", hc.AppFolder, pluginFolder)

	// Create a CA if needed and update launcher and plugin certsclient
	sanNames := []string{hc.Address, "localhost", "127.0.0.1"}
	err = certsetup.CreateCertificateBundle(sanNames, hc.CertsFolder, !hc.KeepServerCertOnStartup)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// start the plugins unless disabled
	if noPlugins {
		logrus.Infof("Starting Hub without plugins")
	} else {
		// launch plugins
		logrus.Infof("Starting %d plugins on %s.", len(pluginConfig.Plugins), hc.Address)

		args := os.Args[1:] // pass the hubs args to the plugin
		StartPlugins(pluginFolder, pluginConfig.Plugins, args)
	}

	logrus.Warningf("Hub start successful!")

	return nil
}

// StopHub stops a running launcher and its plugins
// TODO implements
func StopHub() {
	logrus.Warningf("Received Signal, stopping launcher and its plugins")
	logrus.Warningf("Unable to stop launcher plugins. Someone hasn't implemented this yet...")
}
