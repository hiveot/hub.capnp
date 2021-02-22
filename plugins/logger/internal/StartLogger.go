package internal

import (
	"flag"
	"path"

	"github.com/wostzone/gateway/pkg/lib"
)

const pluginID = "logger"

// StartLogger starts the logger plugin
// This requires a configuration file with the logging channels
func StartLogger() error {
	config := &Config{}
	err := lib.LoadConfigFromCommandline(pluginID+".yaml", config)
	if err != nil {
		return err
	}
	flag.Parse()
	lib.SetLogging(config.Logging.Loglevel, path.Join(config.Logging.LogsFolder, pluginID+".log"))

	pl := NewLoggerPlugin(config)
	pl.Start()
	return nil
}
