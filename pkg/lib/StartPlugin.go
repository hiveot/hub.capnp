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
// This captures the plugin stderr output to report on plugin startup problems
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
