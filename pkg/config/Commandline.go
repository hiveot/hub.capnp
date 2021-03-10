package config

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

var flagsAreSet bool = false

// SetHubCommandlineArgs creates common hub commandline flag commands for parsing commandlines
func SetHubCommandlineArgs(config *HubConfig) {
	// // workaround broken testing of go flags, as they define their own flags that cannot be re-initialized
	// // in test mode this function can be called multiple times. Since flags cannot be
	// // defined multiplte times, prevent redefining them just like testing.Init does.
	// if flagsAreSet {
	// 	return
	// }
	flagsAreSet = true
	// Flags -c and --home are handled separately in SetupConfig. It is added here to avoid flag parse error
	flag.String("c", "hub.yaml", "Set the hub configuration file ")
	flag.StringVar(&config.Home, "home", config.Home, "Application working `folder`")

	flag.StringVar(&config.Messenger.CertFolder, "certFolder", config.Messenger.CertFolder, "Optional certificate `folder` for TLS")
	flag.StringVar(&config.ConfigFolder, "configFolder", config.ConfigFolder, "Plugin configuration `folder`")
	flag.StringVar(&config.Messenger.HostPort, "hostname", config.Messenger.HostPort, "Message bus address host:port")
	flag.StringVar(&config.Logging.LogFile, "logFile", config.Logging.LogFile, "Log to file")
	flag.StringVar(&config.Messenger.Protocol, "protocol", string(config.Messenger.Protocol), "Message bus protocol: internal|mqtt")
	flag.StringVar(&config.PluginFolder, "pluginFolder", config.PluginFolder, "Alternate plugin `folder`. Empty to not load plugins.")
	flag.StringVar(&config.Logging.Loglevel, "logLevel", config.Logging.Loglevel, "Loglevel: {error|`warning`|info|debug}")
}

// SetupConfig contains the boilerplate to load the hub and plugin configuration files.
// parse the commandline and set the plugin logging configuration.
// The caller can add custom commandline options beforehand using the flag package.
//
// The default hub config filename is 'hub.yaml' (const HubConfigName)
// The plugin configuration is the {pluginName}.yaml. If no pluginName is given it is ignored.
// The plugin logfile is stored in the hub logging folder using the pluginName.log filename
// This loads the hub commandline arguments with two special considerations:
//  - Commandline "-c"  specifies an alternative hub configuration file
//  - Commandline "--home" sets the home folder as the base of ./config, ./logs and ./bin directories
//       The homeFolder argument takes precedence
//
// homeFolder overrides the default home folder
//     Leave empty to use parent of application binary. Intended for running tests.
//     The current working directory is changed to this folder
// pluginName is used as the ID in messaging and the plugin configuration filename
//     The plugin config file is optional. Sensible defaults will be used if not present.
// pluginConfig is the configuration to load. nil to only load the hub config.
// Returns the hub configuration and error code in case of error
func SetupConfig(homeFolder string, pluginName string, pluginConfig interface{}) (*HubConfig, error) {
	args := os.Args[1:]
	if homeFolder == "" {
		// Option --home overrides the default home folder. Intended for testing.
		for index, arg := range args {
			if arg == "--home" || arg == "-home" {
				homeFolder = args[index+1]
				// make relative paths absolute
				if !path.IsAbs(homeFolder) {
					cwd, _ := os.Getwd()
					homeFolder = path.Join(cwd, homeFolder)
				}
				break
			}
		}
	}

	// set configuration defaults
	hubConfig := CreateDefaultHubConfig(homeFolder)
	hubConfigFile := path.Join(hubConfig.ConfigFolder, HubConfigName)

	// Option -c overrides the default hub config file. Intended for testing.
	args = os.Args[1:]
	for index, arg := range args {
		if arg == "-c" {
			hubConfigFile = args[index+1]
			// make relative paths absolute
			if !path.IsAbs(hubConfigFile) {
				hubConfigFile = path.Join(homeFolder, hubConfigFile)
			}
			break
		}
	}
	logrus.Infof("Using %s as hub config file", hubConfigFile)
	err1 := LoadConfig(hubConfigFile, hubConfig)
	if err1 != nil {
		// panic("Unable to continue without hub.yaml")
		return hubConfig, err1
	}
	err2 := ValidateConfig(hubConfig)
	if err2 != nil {
		return hubConfig, err2
	}
	if pluginName != "" && pluginConfig != nil {
		pluginConfigFile := path.Join(hubConfig.ConfigFolder, pluginName+".yaml")
		err3 := LoadConfig(pluginConfigFile, pluginConfig)
		if err3 != nil {
			// ignore errors. The plugin configuration file is optional
			// return hubConfig, err3
		}
	}
	SetHubCommandlineArgs(hubConfig)
	// catch parsing errors, in case flag.ErrorHandling = flag.ContinueOnError
	err := flag.CommandLine.Parse(os.Args[1:])

	// Second validation pass in case commandline argument messed up the config
	if err == nil {
		err = ValidateConfig(hubConfig)
		// if err != nil {
		// 	logrus.Errorf("Commandline configuration invalid: %s", err)
		// }
	}

	// It is up to the app to change to the home directory.
	// os.Chdir(hubConfig.HomeFolder)

	// Last set the hub/plugin logging
	if pluginName != "" {
		logFolder := path.Dir(hubConfig.Logging.LogFile)
		logFileName := path.Join(logFolder, pluginName+".log")
		SetLogging(hubConfig.Logging.Loglevel, logFileName, hubConfig.Logging.TimeFormat)
	} else if hubConfig.Logging.LogFile != "" {
		SetLogging(hubConfig.Logging.Loglevel, hubConfig.Logging.LogFile, hubConfig.Logging.TimeFormat)
	}
	return hubConfig, err
}
