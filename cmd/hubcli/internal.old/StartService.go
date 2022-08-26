// Package internal for starting and stopping services using go-cmd
package internal_old

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/sirupsen/logrus"
)

// map of service by ID and PID
var startedServices map[string]*exec.Cmd = make(map[string]*exec.Cmd)

// var servicesMutex = sync.Mutex{}
var servicesMutex = sync.Mutex{}

// BuildDaprCommandline builds the commandline arguments to launch the application with dapr sidecar
//
// config is the configuration will all necessary parameters filled in
// This returns the array of commandline arguments starting with 'dapr run'
func BuildDaprCommandline(config ServiceConfig) (args []string) {
	args = []string{"dapr", "run"}
	if config.SvcID != "" {
		args = append(args, "--app-id", config.SvcID)
	}

	if config.DaprHttpPort != 0 {
		args = append(args, "--dapr-http-port", strconv.Itoa(config.DaprHttpPort))
	}
	if config.DaprGrpcPort != 0 {
		args = append(args, "--dapr-grpc-port", strconv.Itoa(config.DaprGrpcPort))
	}
	if config.DaprGrpcPort != 0 {
		args = append(args, "--dapr-grpc-port", strconv.Itoa(config.DaprGrpcPort))
	}
	if config.LogLevel != "" {
		args = append(args, "--log-level", config.LogLevel)
	}
	if config.ServiceSocket != "" {
		args = append(args, "--unix-domain-socket", config.ServiceSocket)
	} else if config.ServicePort != 0 {
		args = append(args, "--app-port", strconv.Itoa(config.ServiceAppPort))
	}
	return
}

// ListServices lists the status of the services in the given configuration
// This returns an array [name, autostart, status, duration]
func ListServices(config LauncherConfig) {

}

// StartService starts the service using the given configuration
// This returns the running command, or an error if start fails
func StartService(service ServiceConfig) *cmd.Cmd {
	commandline := []string{service.AppBin}
	cmd_1 := fmt.Sprintf("dapr run %s --app-id %s --log-level %s", service.AppBin, service.AppID, service.LogLevel)
	launchCmd := cmd.NewCmd(service.AppBin, commandline)
	return launchCmd
}

// StopService stops the started service with the given name
func StopService(service ServiceConfig) {

}

// StartService_old runs the executable with the given name.
// If the name contains a relative path, it is appended to the pluginFolder
// If the name contains an absolute path the pluginFolder is ignored.
// See https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
//  for some good tips on executing commands with osexec.
// This captures the plugin stderr output to report on plugin startup problems
//
//  pluginFolder is specific folder for plugins. Leave empty to search the OS PATH
//  name is the executable name of the plugin.
//  args is a list of commandline arguments
// returns the plugin command
func StartService_old(pluginFolder string, name string, args []string) *exec.Cmd {
	// logrus.Warningf("--: '%s'---", name)
	logrus.Warningf("----------- Starting plugin '%s' ------------", name)
	pluginFile := name
	if !filepath.IsAbs(name) {
		pluginFile = path.Join(pluginFolder, name)
	}

	servicesMutex.Lock()
	defer servicesMutex.Unlock()
	// If plugin is already running, don't start it twice
	// TODO: does it need to support multiple instances?
	exists := startedServices[name]
	if exists != nil {
		// TODO: check if process is running
		logrus.Errorf("StartService: Plugin with name %s has already been started", name)
		return nil
	}
	// argString := strings.Join(args, " ")
	cmd := exec.Command(pluginFile, args...)
	// Capture stderr in case of startup failure
	cmd.Stderr = os.Stderr
	// keep track of what is started. This doesn't mean it is running though
	startedServices[name] = cmd
	err := cmd.Start()
	if err != nil {
		logrus.Errorf("Plugin '%s' ended with error: %s", name, err)
		return nil
	}

	go func() {
		err = cmd.Wait()
		if err != nil {
			logrus.Errorf("Plugin '%s' ended with error: %s", name, err)
		} else {
			logrus.Warningf("Plugin '%s' has ended", name)
		}
		servicesMutex.Lock()
		startedServices[name] = nil
		defer servicesMutex.Unlock()
	}()
	// logrus.Errorf("StartService '%s'", name)

	// cmd.Stdout = os.Stdout
	// logrus.Warningf("----------- Started plugin '%s' ------------", name)
	// Give room to switch threads
	time.Sleep(time.Millisecond)
	return cmd
}

// StartAllServices starts all services in launcher.yaml with autostart enabled
func StartAllServices(pluginFolder string, names []string, args []string) {
	for _, name := range names {
		StartService(pluginFolder, name, args)
	}
}

// StopService stops the service with the given name
func StopService_old(name string) error {
	servicesMutex.Lock()
	defer servicesMutex.Unlock()

	cmd := startedServices[name]
	if cmd == nil || cmd.Process == nil {
		msg := fmt.Sprintf("Failed to stop plugin '%s'. Plugin not running", name)
		logrus.Errorf(msg)
		return errors.New(msg)
	}
	logrus.Warningf("Stopped plugin '%s', PID='%d'", name, cmd.Process.Pid)
	err := syscall.Kill(cmd.Process.Pid, syscall.SIGINT)
	startedServices[name] = nil
	return err
}

// StopAllServices stops all started services
func StopAllServices() {
	servicesMutex.Lock()
	keys := make([]string, 0)
	for key := range startedServices {
		keys =
			append(keys, key)
	}
	servicesMutex.Unlock()

	for _, key := range keys {
		StopService(key)
	}
}
