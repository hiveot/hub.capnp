package lib

import (
	"flag"
	"strings"
)

// SetupGatewayArgs creates common gateway and plugin commandline arguments
// for parsing commandlines
func SetupGatewayArgs(config *GatewayConfig) {
	flag.StringVar(&config.Messenger.CertsFolder, "certsFolder", config.Messenger.CertsFolder, "Optional certificate `folder` for TLS")
	flag.StringVar(&config.ConfigFolder, "configFolder", config.ConfigFolder, "Gateway and plugin configuration `folder`")
	flag.StringVar(&config.Messenger.HostPort, "hostname", config.Messenger.HostPort, "Message bus address host:port")
	flag.StringVar(&config.Logging.LogsFolder, "logsFolder", config.Logging.LogsFolder, "Optional logs `folder`")
	flag.StringVar(&config.Messenger.Protocol, "protocol", config.Messenger.Protocol, "Message bus protocol: internal|mqtt")
	flag.StringVar(&config.PluginFolder, "pluginFolder", config.PluginFolder, "Optional plugin `folder`")
	flag.StringVar(&config.Logging.Loglevel, "loglevel", config.Logging.Loglevel, "Loglevel: {error|`warning`|info|debug}")
	flag.BoolVar(&config.Messenger.UseTLS, "useTLS", config.Messenger.UseTLS, "Gateway listens using TLS {`true`|false}")
}

// ParseCommandline parses the given commandline arguments
// flag must have been setup, for example by calling SetupGatewayArgs()
func ParseCommandline(cmdArgs string, config interface{}) error {
	myArgList := strings.Split(cmdArgs, " ")
	err := flag.CommandLine.Parse(myArgList)
	return err
}
