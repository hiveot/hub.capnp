package config

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// LoadAllConfig is a helper to load all configuration from commandline, hubconfig and client config
// This:
//  1. Determine application defaults
//  2. parse commandline arguments for options -c hub.yaml -a appFolder or -h
//  3. Load the hub global configuration file hub.yaml, if found
//  4. Load the client configuration file {clientID}.yaml, if found
//
//  args is the os.argv list. Use nil to ignore commandline args
//  appFolder is the installation folder, "" for default parent folder of app binary
//  clientID is the server, plugin or device instance ID. Used when connecting to servers
//  clientConfig is an instance of the client's configuration object
// This returns the hub global configuration with an error if something went wrong
func LoadAllConfig(args []string, appFolder string, clientID string, clientConfig interface{}) (*HubConfig, error) {
	hubConfigFile := DefaultHubConfigName

	// Determine the default application installation folder
	if appFolder == "" {
		appBin, _ := os.Executable()
		binFolder := path.Dir(appBin)
		appFolder = path.Dir(binFolder)
	}

	// Parse commandline arguments for options -c configFile and -a appFolder
	if args != nil {
		var cmdHubConfigFile string
		var cmdAppFolder string
		flag.StringVar(&cmdHubConfigFile, "c", "", "Change the global hub configuration file")
		flag.StringVar(&cmdAppFolder, "a", "", "Change the application home folder with config and cert subfolders")
		flag.Parse()
		// override the application home folder
		if cmdAppFolder != "" {
			// relative path is to current working directory, not the app folder
			if cmdHubConfigFile != "" && !path.IsAbs(cmdHubConfigFile) {
				cwd, _ := os.Getwd()
				appFolder = path.Join(cwd, cmdAppFolder)
			} else {
				appFolder = cmdAppFolder
			}
		}
		// manually running config from a different location
		if cmdHubConfigFile != "" {
			// relative path is to current working directory, not the app folder
			if !path.IsAbs(cmdHubConfigFile) {
				cwd, _ := os.Getwd()
				hubConfigFile = path.Join(cwd, cmdHubConfigFile)
			} else {
				hubConfigFile = cmdHubConfigFile
			}
		}
	}
	hubConfig := CreateDefaultHubConfig(appFolder)

	if !path.IsAbs(hubConfigFile) {
		hubConfigFile = path.Join(hubConfig.ConfigFolder, hubConfigFile)
	}
	// logrus.Infof("Set hub config file to %s", hubConfigFile)
	// hub.yaml must exist to continue
	if _, err := os.Stat(hubConfigFile); err != nil {
		logrus.Errorf("LoadConfig: Global hub config not found: %s", err)
		return hubConfig, err
	}
	// load the global hub configuration with certificates
	err := LoadHubConfig(hubConfigFile, clientID, hubConfig)
	if err != nil {
		logrus.Errorf("LoadConfig: Global hub config failed to load: %s", err)
		return hubConfig, err
	}
	// last the client settings (optional)
	if clientConfig != nil {
		clientConfigFile := path.Join(hubConfig.ConfigFolder, clientID+".yaml")
		substituteMap := make(map[string]string)
		substituteMap["{clientID}"] = clientID
		substituteMap["{appFolder}"] = hubConfig.AppFolder
		substituteMap["{configFolder}"] = hubConfig.ConfigFolder
		substituteMap["{logsFolder}"] = hubConfig.LogsFolder
		substituteMap["{certsFolder}"] = hubConfig.CertsFolder

		if _, err = os.Stat(clientConfigFile); os.IsNotExist(err) {
			logrus.Infof("FYI The optional client configuration file %s is not present", clientConfigFile)
			err = nil
		} else {
			err = LoadYamlConfig(clientConfigFile, clientConfig, substituteMap)
		}
	}
	return hubConfig, err
}
