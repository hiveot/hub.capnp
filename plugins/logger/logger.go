package logger

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/sirupsen/logrus"
	messenger "github.com/wostzone/gateway/src/messenger/go"
	"gopkg.in/yaml.v2"
)

// PluginID is the ID of the logger plugin
const PluginID = "logger"

// Config with logging configuration, default channels logged are td, events, actions
type Config struct {
	Channels   []string `yaml:"channels"`
	LogsFolder string   `yaml:"logsFolder"` // default is ../logs
	UseTLS     bool     `yaml:"useTLS`
	Loglevel   string   `yaml:"loglevel"`
}

// Set default configuration and load optional configuration file
func loadConfig(configFile string) *Config {
	gwbin, _ := os.Executable()
	binFolder := path.Dir(gwbin)
	appFolder := path.Dir(binFolder)
	// for running within the project use the test folder as application root folder
	if path.Base(binFolder) != "bin" {
		appFolder = path.Join(appFolder, "test")
	}
	config := &Config{
		Channels:   []string{messenger.TDChannelID, messenger.ActionChannelID, messenger.EventsChannelID},
		LogsFolder: path.Join(appFolder, "logs"),
	}

	// configFile := path.Join(config.ConfigFolder, "gateway.yaml")
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

// StartPlugin starts the logging plugin
func StartPlugin() {
	var hostPort string
	var configFile string
	var certFolder string
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		println("Usage: plugin host [configFile [certFolder]] ")
		return
	}
	hostPort = args[0]
	if len(args) == 2 {
		configFile = args[1]
	}
	if len(args) == 3 {
		configFile = args[1]
		certFolder = args[2]
	}
	config := loadConfig(configFile)
	messenger.SetLogging(config.Loglevel, path.Join(config.LogsFolder, PluginID+".log"))
	plugin := NewLoggingPlugin(hostPort, certFolder)
	plugin.Start()
	// wait for signal to end
	waitForSignal()
	plugin.Stop()
}

// import gateway src/gateway
// arguments: host, configFile, certFolder
func main() {

	StartPlugin()

}
