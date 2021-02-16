// Package gateway with gateway source
package gateway

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/src/msgbus"
)

const hostname = "localhost:9678"

// GatewayConfig with commandline configuration parameters
type GatewayConfig struct {
	loglevel   string // debug, info, warning, error. Default is warning
	logs       string // location of log files. Default
	messagebus string // internal, MQTT, default internal
	tls        bool   // use TLS for client/server messaging
	hostname   string // hostname:port or ip:port to listen on

	appFolder     string   // location of application
	configFolder  string   // location of configuration files
	logsFolder    string   // location of logfiles
	pluginsFolder string   // location of plugin binaries
	certsFolder   string   // location of gateway and client certificate
	pluginNames   []string // list of plugins to start
}

// parseArgs parses commandline arguments and return the configuration
func parseArgs() GatewayConfig {
	gwbin, _ := os.Executable()
	// appFolder := path.Dir(path.Dir(gwbin))
	appFolder := path.Dir(gwbin)
	if path.Base(appFolder) == "src" {
		appFolder = path.Join(appFolder, "../test")
	}
	// if this is the src folder, then this is running from the debugger, change the app folder
	// to the test folder

	config := GatewayConfig{
		configFolder:  path.Join(appFolder, "/config"),
		hostname:      "localhost:9678",
		logsFolder:    path.Join(appFolder, "logs"),
		messagebus:    "",
		certsFolder:   path.Join(appFolder, "/certs"),
		pluginNames:   make([]string, 0),
		pluginsFolder: path.Join(appFolder, "/plugins"),
	}

	flag.StringVar(&config.certsFolder, "certs", config.certsFolder, "Optional certificate `folder` for TLS")
	flag.StringVar(&config.configFolder, "config", config.configFolder, "Gateway and plugin configuration `folder`")
	flag.StringVar(&config.hostname, "hostname", config.hostname, "Message bus address")
	flag.StringVar(&config.logsFolder, "logs", config.logsFolder, "Optional logs `folder`")
	flag.StringVar(&config.pluginsFolder, "plugins", config.pluginsFolder, "Optional plugin `folder`")
	flag.StringVar(&config.loglevel, "loglevel", "warning", "Loglevel: {error|`warning`|info|debug}")
	flag.BoolVar(&config.tls, "tls", true, "Gateway listens using TLS {`true`|false}")
	// helpPtr := flag.Bool("help", false, "Show help")
	flag.Parse()
	return config
}

// SetLogging sets the logging level and output file for this publisher
// Intended for setting logging from configuration
//  levelName is the requested logging level: error, warning, info, debug
//  filename is the output log file full name including path, use "" for stderr
func setLogging(levelName string, filename string) error {
	loggingLevel := logrus.DebugLevel
	var err error

	if levelName != "" {
		switch strings.ToLower(levelName) {
		case "error":
			loggingLevel = logrus.ErrorLevel
		case "warn":
		case "warning":
			loggingLevel = logrus.WarnLevel
		case "info":
			loggingLevel = logrus.InfoLevel
		case "debug":
			loggingLevel = logrus.DebugLevel
		}
	}
	logOut := os.Stderr
	if filename != "" {
		logFileHandle, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			err = fmt.Errorf("Publisher.SetLogging: Unable to open logfile: %s", err)
		} else {
			logrus.Warnf("Publisher.SetLogging: Send logging output to %s", filename)
			logOut = logFileHandle
		}
	}

	logrus.SetFormatter(
		&logrus.TextFormatter{
			// LogFormat: "",
			// DisableColors:   true,
			// DisableLevelTruncation: true,
			// PadLevelText:    true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			FullTimestamp:   true,
			// ForceFormatting: true,
		})
	logrus.SetOutput(logOut)
	logrus.SetLevel(loggingLevel)

	logrus.SetReportCaller(false) // publisher logging includes caller and file:line#
	return err
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
// Configuration options are:
// --msgbus=internal|mqtt, default is internal
// --host=host:port, default is localhost:9678
// --config=folder, with configuration files, default is $appfolder/config
// --logs=folder, with logging files, default is $appfolder/logs
// --plugins=folder, with plugin binaries, default is $appfolder/bin
// --certs=folder, with CA, server and client certificates to enable TLS
//                 default is $appfolder/tls
// --tls, enable TLS on message bus and plugin messaging client
// --help show help and exit
func StartGateway() {
	config := parseArgs()
	var server *msgbus.ServeMsgBus
	var serverHostPort = msgbus.DefaultMsgBusHost
	var err error

	// read config
	if _, err := os.Stat(config.configFolder); os.IsNotExist(err) {
		logrus.Errorf("Configuration folder '%s' not found\n", config.configFolder)
		os.Exit(1)
	}
	// logrus.Warnf("RunAdapter from config %s", *configFolderPtr)
	// err = LoadConfiguration(*configFolderPtr)
	// if err != nil {
	// 	os.Exit(2)
	// }

	// if *helpPtr {
	// 	os.Exit(0)
	// }

	setLogging(config.loglevel, path.Join(config.logs, "gateway"))
	// determine service bus to use

	// launch internal service bus

	logrus.Warningf("Starting the WoST Gateway on %s", serverHostPort)
	if !config.tls {
		server, err = msgbus.Start(serverHostPort)
	} else {
		server, err = msgbus.StartTLS(serverHostPort, config.certsFolder)
	}
	if err != nil {
		logrus.Errorf("Error starting the internal message bus: ", err)
	}
	// launch plugins
	for _, pluginName := range config.pluginNames {
		startPlugin(pluginName, config.pluginsFolder)
	}

	// wait for signal to end
	waitForSignal()
	server.Stop()
}
