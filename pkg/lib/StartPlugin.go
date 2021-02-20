package lib

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// StartPlugin prepares executing the plugin with the given name.
// If the name contains a relative path, it is appended to the pluginFolder
// If the name contains an absolute path the pluginFolder is ignored.
// See https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
//  for some good tips on executing commands with osexec.
// This captures the command stderr
//
//  pluginFolder is specific folder for plugins. Leave empty to search the OS PATH
//  name is the executable name of the plugin.
//  args is a list of commandline arguments
//  returns exec.Cmd. Execute command with .Run()
func StartPlugin(pluginFolder string, name string, args []string) *exec.Cmd {
	pluginFile := name
	if !filepath.IsAbs(name) {
		pluginFile = path.Join(pluginFolder, name)
	}
	cmd := exec.Command(pluginFile, args...)
	cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	return cmd
}

// StartPlugins starts all plugin names in the pluginFolder
func StartPlugins(pluginFolder string, names []string, args []string) {
	for _, name := range names {
		cmd := StartPlugin(pluginFolder, name, args)
		err := cmd.Run()
		if err != nil {
			logrus.Errorf("StartPlugins: Plugin '%s' failed to start: %s", name, err)
		} else {
			logrus.Warningf("StartPlugins: Started plugin '%s'", name)
		}
	}
}

// // StartPlugin starts the logging plugin
// func StartPlugin() {
// 	var hostPort string
// 	var configFile string
// 	var certFolder string
// 	flag.Parse()

// 	args := flag.Args()
// 	if len(args) < 2 {
// 		println("Usage: plugin host [configFile [certFolder]] ")
// 		return
// 	}
// 	hostPort = args[0]
// 	if len(args) == 2 {
// 		configFile = args[1]
// 	}
// 	if len(args) == 3 {
// 		configFile = args[1]
// 		certFolder = args[2]
// 	}
// 	config := loadConfig(configFile)
// 	logging.SetLogging(config.Loglevel, path.Join(config.LogsFolder, PluginID+".log"))
// 	plugin := NewLoggerPlugin(hostPort, certFolder)
// 	plugin.Start()
// 	// wait for signal to end
// 	waitForSignal()
// 	plugin.Stop()
// }
