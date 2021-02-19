package logger

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
func NewLoggerPlugin(hostPort string, certFolders string) *LoggingPlugin {
	return &LoggingPlugin{}
}
