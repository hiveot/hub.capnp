package internal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// map of plugins by ID and PID
var startedPlugins map[string]*exec.Cmd = make(map[string]*exec.Cmd)

// var pluginsMutex = sync.Mutex{}
var pluginsMutex = sync.Mutex{}

// StartPlugin runs the executable with the given name.
// If the name contains a relative path, it is appended to the pluginFolder
// If the name contains an absolute path the pluginFolder is ignored.
// See https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
//  for some good tips on executing commands with osexec.
// This captures the plugin stderr output to report on plugin startup problems
//
//  pluginFolder is specific folder for plugins. Leave empty to search the OS PATH
//  name is the executable name of the plugin.
//  args is a list of commandline arguments
//  returns *exec.Cmd
func StartPlugin(pluginFolder string, name string, args []string) *exec.Cmd {
	// logrus.Warningf("--: '%s'---", name)
	logrus.Warningf("----------- Starting plugin '%s' ------------", name)
	pluginFile := name
	if !filepath.IsAbs(name) {
		pluginFile = path.Join(pluginFolder, name)
	}

	pluginsMutex.Lock()
	defer pluginsMutex.Unlock()
	// If plugin is already running, don't start it twice
	// TODO: does it need to support multiple instances?
	exists := startedPlugins[name]
	if exists != nil {
		// TODO: check if process is running
		logrus.Errorf("StartPlugin: Plugin with name %s has already been started", name)
		return nil
	}
	// argString := strings.Join(args, " ")
	cmd := exec.Command(pluginFile, args...)
	// Capture stderr in case of startup failure
	cmd.Stderr = os.Stderr
	// keep track of what is started. This doesn't mean it is running though
	startedPlugins[name] = cmd
	err := cmd.Start()
	if err != nil {
		logrus.Errorf("StartPlugin: Plugin '%s' ended with error: %s", name, err)
		return nil
	}

	go func() {
		err = cmd.Wait()
		if err != nil {
			logrus.Errorf("StartPlugin: Plugin '%s' ended with error: %s", name, err)
		} else {
			logrus.Warningf("StartPlugin: Plugin '%s' has ended", name)
		}
		pluginsMutex.Lock()
		startedPlugins[name] = nil
		defer pluginsMutex.Unlock()
	}()
	// logrus.Errorf("StartPlugin '%s'", name)

	// cmd.Stdout = os.Stdout
	// logrus.Warningf("----------- Started plugin '%s' ------------", name)
	// Give room to switch threads
	time.Sleep(time.Millisecond)
	return cmd
}

// StartPlugins starts all plugin names in the pluginFolder
func StartPlugins(pluginFolder string, names []string, args []string) {
	for _, name := range names {
		StartPlugin(pluginFolder, name, args)
	}
}

// StopPlugin stops a plugin by name
func StopPlugin(name string) error {
	pluginsMutex.Lock()
	defer pluginsMutex.Unlock()

	cmd := startedPlugins[name]
	if cmd == nil || cmd.Process == nil {
		msg := fmt.Sprintf("Failed to stop plugin '%s'. Plugin not running", name)
		logrus.Errorf(msg)
		return errors.New(msg)
	}
	logrus.Warningf("Stopped plugin '%s', PID='%d'", name, cmd.Process.Pid)
	err := syscall.Kill(cmd.Process.Pid, syscall.SIGINT)
	startedPlugins[name] = nil
	return err
}

// StopAllPlugins stops all started plugins
func StopAllPlugins() {
	pluginsMutex.Lock()
	keys := make([]string, 0)
	for key := range startedPlugins {
		keys =
			append(keys, key)
	}
	pluginsMutex.Unlock()

	for _, key := range keys {
		StopPlugin(key)
	}
}
