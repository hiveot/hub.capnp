package internal_old

import (
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

const PluginID = "launcher"

// PluginConfig with list of plugins to launch
type PluginConfig struct {
	Plugins []string `yaml:"plugins"`
}

// StartHub reads the launcher configuration and launches the plugins. If the configuration is invalid
// then start is aborted.
//
//  homeFolder is the parent folder of the application binary and contains the config, certs and log subfolders.
//  startPlugins set to false to only start the launcher for testing
//
// Return nil or error if the launcher configuration file or certificate are not found
func StartHub(homeFolder string, startPlugins bool) error {

	var err error
	var noPlugins bool
	var pluginConfig PluginConfig
	var address = "localhost"

	//logging.SetLogging(hc.Loglevel, hc.LogFile)
	pluginFolder := path.Join(homeFolder, "bin")
	fmt.Printf("Home=%s\nPluginFolder=%s\n", homeFolder, pluginFolder)

	// Create a CA if needed and update launcher and plugin certs
	//sanNames := []string{hc.Address, "localhost", "127.0.0.1"}
	//err = certsetup.CreateCertificateBundle(sanNames, hc.CertsFolder, !hc.KeepServerCertOnStartup)
	//if err != nil {
	//	logrus.Error(err)
	//	return err
	//}

	// start the plugins unless disabled
	if noPlugins {
		logrus.Infof("Starting Hub without plugins")
	} else {
		// launch plugins
		logrus.Infof("Starting %d plugins on %s.", len(pluginConfig.Plugins), address)

		args := os.Args[1:] // pass the hubs args to the plugin
		StartAllServices(pluginFolder, pluginConfig.Plugins, args)
	}

	logrus.Warningf("Hub start successful!")

	return err
}

// StopHub stops a running launcher and its plugins
// TODO implements
func StopHub() {
	logrus.Warningf("Received Signal, stopping launcher and its plugins")
	logrus.Warningf("Unable to stop launcher plugins. Someone hasn't implemented this yet...")
}
