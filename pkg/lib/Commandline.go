package lib

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

var flagsAreSet bool = false

// SetGatewayCommandlineArgs creates common gateway commandline flag commands for parsing commandlines
func SetGatewayCommandlineArgs(config *GatewayConfig) {
	// // workaround broken testing of go flags, as they define their own flags that cannot be re-initialized
	// // in test mode this function can be called multiple times. Since flags cannot be
	// // defined multiplte times, prevent redefining them just like testing.Init does.
	// if flagsAreSet {
	// 	return
	// }
	flagsAreSet = true
	// Flags -c and --home are handled separately in SetupConfig. It is added here to avoid flag parse error
	flag.String("c", "gateway.yaml", "Set the gateway configuration file ")
	flag.StringVar(&config.Home, "home", config.Home, "Application working `folder`")

	flag.StringVar(&config.Messenger.CertsFolder, "certsFolder", config.Messenger.CertsFolder, "Optional certificate `folder` for TLS")
	flag.StringVar(&config.ConfigFolder, "configFolder", config.ConfigFolder, "Plugin configuration `folder`")
	flag.StringVar(&config.Messenger.HostPort, "hostname", config.Messenger.HostPort, "Message bus address host:port")
	flag.StringVar(&config.Logging.LogFile, "logFile", config.Logging.LogFile, "Log to file")
	flag.StringVar(&config.Messenger.Protocol, "protocol", string(config.Messenger.Protocol), "Message bus protocol: internal|mqtt")
	flag.StringVar(&config.PluginFolder, "pluginFolder", config.PluginFolder, "Optional plugin `folder`")
	flag.StringVar(&config.Logging.Loglevel, "logLevel", config.Logging.Loglevel, "Loglevel: {error|`warning`|info|debug}")
	flag.BoolVar(&config.Messenger.UseTLS, "useTLS", config.Messenger.UseTLS, "Gateway listens using TLS {`true`|false}")
}

// SetupConfig contains the boilerplate to load the gateway and plugin configuration files.
// parse the commandline and set the plugin logging configuration.
// The plugin should add any custom commandline options with the flag package before calling SetupConfig.
//
// The gateway config filename is always GatewayConfigName ('gateway.yaml')
// The plugin configuration is the {pluginName}.yaml. If no pluginName is given it is ignored
//  and logging for the plugin is not configured.
// The plugin logfile is stored in the gateway logging folder using the pluginName.log filename
// This loads the gateway commandline arguments with two special considerations:
//  - Option "-c"  loads the specified configuration file instead of the default one
//  - Option "--home" sets the home folder as the base of ./config, ./logs and ./bin directories
//       The homeFolder argument takes precedence
//
// homeFolder overrides the default home folder
//     Leave empty to use parent of application binary. Intended for running tests.
//     The current working directory is changed to this folder
// pluginName is used as the ID in messaging and the plugin configuration filename
//     The plugin config file is optional. Sensible defaults will be used if not present.
// pluginConfig is the configuration to load. nil to only load the gateway config.
// Returns the gateway configuration and error code in case of error
func SetupConfig(homeFolder string, pluginName string, pluginConfig interface{}) (*GatewayConfig, error) {
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
	gwConfig := CreateDefaultGatewayConfig(homeFolder)
	gwConfigFile := path.Join(gwConfig.ConfigFolder, GatewayConfigName)

	// Option -c overrides the default gateway config file. Intended for testing.
	args = os.Args[1:]
	for index, arg := range args {
		if arg == "-c" {
			gwConfigFile = args[index+1]
			// make relative paths absolute
			if !path.IsAbs(gwConfigFile) {
				gwConfigFile = path.Join(homeFolder, gwConfigFile)
			}
			break
		}
	}
	logrus.Infof("SetupConfig: Using %s as gateway config file", gwConfigFile)
	err1 := LoadConfig(gwConfigFile, gwConfig)
	if err1 != nil {
		return gwConfig, err1
	}
	err2 := ValidateConfig(gwConfig)
	if err2 != nil {
		return gwConfig, err2
	}
	if pluginName != "" && pluginConfig != nil {
		pluginConfigFile := path.Join(gwConfig.ConfigFolder, pluginName+".yaml")
		err3 := LoadConfig(pluginConfigFile, pluginConfig)
		if err3 != nil {
			// ignore errors. The plugin configuration file is optional
			// return gwConfig, err3
		}
	}
	SetGatewayCommandlineArgs(gwConfig)
	// catch parsing errors, in case flag.ErrorHandling = flag.ContinueOnError
	err := flag.CommandLine.Parse(os.Args[1:])

	// Second validation pass in case commandline argument messed up the config
	if err == nil {
		err = ValidateConfig(gwConfig)
		if err != nil {
			logrus.Errorf("SetupConfig: commandline configuration invalid: %s", err)
		}
	}

	// It is up to the app to change to the home directory.
	// os.Chdir(gwConfig.HomeFolder)

	// Last set the gateway/plugin logging
	if pluginName != "" {
		logFolder := path.Dir(gwConfig.Logging.LogFile)
		logFileName := path.Join(logFolder, pluginName+".log")
		SetLogging(gwConfig.Logging.Loglevel, logFileName)
	} else if gwConfig.Logging.LogFile != "" {
		SetLogging(gwConfig.Logging.Loglevel, gwConfig.Logging.LogFile)
	}
	return gwConfig, err
}
