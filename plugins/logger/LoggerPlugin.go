package logger

// PluginID is the ID of the logger plugin
const PluginID = "logger"

// Config with logging configuration, default channels logged are td, events, actions
type Config struct {
	Channels   []string `yaml:"channels"`
	LogsFolder string   `yaml:"logsFolder"` // default is ../logs
	UseTLS     bool     `yaml:"useTLS"`
	Loglevel   string   `yaml:"loglevel"`
}

// LoggerPlugin is a gateway plugin for logging channels
type LoggerPlugin struct {
}

// Start the logging
func (plugin *LoggerPlugin) Start() {

}

// Stop the logging
func (plugin *LoggerPlugin) Stop() {

}

// NewLoggerPlugin creates a new logging plugin instance
func NewLoggerPlugin(hostPort string, certFolders string) *LoggerPlugin {
	return &LoggerPlugin{}
}
