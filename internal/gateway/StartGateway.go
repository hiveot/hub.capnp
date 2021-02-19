// Package gateway with gateway source
package gateway

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/logging"
	"gopkg.in/yaml.v2"
)

const hostname = "localhost:9678"

// AppConfig with gateway configuration parameters
type AppConfig struct {
	AppFolder    string   `yaml:"app"`          // application root folder
	CertsFolder  string   `yaml:"certsFolder"`  // location of gateway and client certificate
	ConfigFolder string   `yaml:"configFolder"` // location of plugin configuration files
	HostPort     string   `yaml:"hostname"`     // hostname:port or ip:port to listen on of message bus
	Loglevel     string   `yaml:"loglevel"`     // debug, info, warning, error. Default is warning
	LogsFolder   string   `yaml:"logsFolder"`   // location of logfiles
	Protocol     string   `yaml:"protocol"`     // internal, MQTT, default internal
	PluginFolder string   `yaml:"pluginFolder"` // location of plugin binaries
	PluginNames  []string `yaml:"pluginNames"`  // names of plugins to start
	UseTLS       bool     `yaml:"useTLS"`       // use TLS for client/server messaging

	// internal
}

// parseArgs parses commandline arguments and update configuration
func parseArgs(config *AppConfig) {

	flag.StringVar(&config.CertsFolder, "certsFolder", config.CertsFolder, "Optional certificate `folder` for TLS")
	flag.StringVar(&config.ConfigFolder, "configFolder", config.ConfigFolder, "Gateway and plugin configuration `folder`")
	flag.StringVar(&config.HostPort, "hostname", config.HostPort, "Message bus address host:port")
	flag.StringVar(&config.LogsFolder, "logsFolder", config.LogsFolder, "Optional logs `folder`")
	flag.StringVar(&config.Protocol, "protocol", config.Protocol, "Message bus protocol: internal|mqtt")
	flag.StringVar(&config.PluginFolder, "pluginFolder", config.PluginFolder, "Optional plugin `folder`")
	flag.StringVar(&config.Loglevel, "loglevel", config.Loglevel, "Loglevel: {error|`warning`|info|debug}")
	flag.BoolVar(&config.UseTLS, "useTLS", config.UseTLS, "Gateway listens using TLS {`true`|false}")
	flag.Parse()

	pluginNames := flag.Args()
	if len(pluginNames) > 0 {
		config.PluginNames = pluginNames
	}
	// helpPtr := flag.Bool("help", false, "Show help")
}

// Set default configuration and load optional configuration file
func loadConfig() *AppConfig {
	gwbin, _ := os.Executable()
	binFolder := path.Dir(gwbin)
	appFolder := path.Dir(binFolder)
	// for running within the project use the test folder as application root folder
	if path.Base(binFolder) != "bin" {
		appFolder = path.Join(appFolder, "test")
	}
	// if this is the src folder, then this is running from the debugger, change the app folder
	// to the test folder

	config := &AppConfig{
		ConfigFolder: path.Join(appFolder, "config"),
		HostPort:     "localhost:9678",
		LogsFolder:   path.Join(appFolder, "logs"),
		Protocol:     "",
		CertsFolder:  path.Join(appFolder, "certs"),
		PluginNames:  make([]string, 0),
		PluginFolder: path.Join(appFolder, "bin"),
		UseTLS:       true,
	}
	configFile := path.Join(config.ConfigFolder, "gateway.yaml")
	rawConfig, err := ioutil.ReadFile(configFile)
	if err == nil {
		logrus.Infof("Loading configuration from: %s", configFile)
		err = yaml.Unmarshal(rawConfig, config)
		if err != nil {
			logrus.Errorf("Failed parsing configuration file %s: %s", configFile, err)
		}
	}
	return config
}

// WaitForSignal waits until a TERM or INT signal is received
func waitForSignal() {

	// catch all signals since not explicitly listing
	exitChannel := make(chan os.Signal, 1)

	//signal.Notify(exitChannel, syscall.SIGTERM|syscall.SIGHUP|syscall.SIGINT)
	signal.Notify(exitChannel, syscall.SIGINT, syscall.SIGTERM)

	sig := <-exitChannel
	logrus.Warningf("RECEIVED SIGNAL: %s", sig)
	fmt.Println()
	fmt.Println(sig)
}

// startPlugin start the plugin with the given name
func startPlugin(name string, folder string) {
	logrus.Warningf("TODO: plugins not yet supported")
}

// StartGateway launches the gateway and its plugins
// By default this uses the internal message bus without TLS
// see parseArgs for commandline options
// This loads the optional gateway configuration file 'gateway.yaml'
func StartGateway() {
	// var server *msgbus.ServeMsgBus
	// var serverHostPort = msgbus.DefaultMsgBusHost
	var err error

	config := loadConfig()
	parseArgs(config)

	// check folders exist
	if _, err := os.Stat(config.ConfigFolder); os.IsNotExist(err) {
		logrus.Errorf("Configuration folder '%s' not found\n", config.ConfigFolder)
		os.Exit(1)
	} else if _, err := os.Stat(config.LogsFolder); os.IsNotExist(err) {
		logrus.Errorf("Logging folder '%s' not found\n", config.LogsFolder)
		os.Exit(1)
	} else if _, err := os.Stat(config.CertsFolder); os.IsNotExist(err) && config.UseTLS {
		logrus.Errorf("TLS certificate folder '%s' not found\n", config.CertsFolder)
		os.Exit(1)
	} else if _, err := os.Stat(config.PluginFolder); os.IsNotExist(err) {
		logrus.Errorf("Plugins folder '%s' not found\n", config.PluginFolder)
		// os.Exit(1)
	}

	logging.SetLogging(config.Loglevel, path.Join(config.LogsFolder, "gateway.log"))

	// logrus.Warningf("Starting the WoST Gateway on %s. UseTLS=%t", serverHostPort, config.UseTLS)
	logrus.Infof("Starting %d plugin(s)", len(config.PluginNames))
	// os.Exit(0)
	// if !config.UseTLS {
	// 	server, err = msgbus.Start(serverHostPort)
	// } else {
	// 	server, err = msgbus.StartTLS(serverHostPort, config.CertsFolder)
	// }
	if err != nil {
		logrus.Errorf("Error starting the internal message bus: %s", err)
	}

	// launch plugins
	logrus.Warningf("Starting %d plugins", len(config.PluginNames))
	for _, pluginName := range config.PluginNames {
		startPlugin(pluginName, config.PluginFolder)
	}

	// wait for signal to end
	waitForSignal()
	// server.Stop()
}
