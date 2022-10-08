package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/launcher"
)

// LauncherService manages starting and stopping of services
// This implements the ILauncher interface
type LauncherService struct {
	// map of service name to running status
	services map[string]*launcher.ServiceInfo
	// map of started commands
	cmds map[string]*cmd.Cmd
	// mutex to keep things safe
	mux sync.Mutex
}

// Add newly discovered executable services
// If the service is already know, only update its size and timestamp
func (ls *LauncherService) findServices(folder string) error {

	entries, err := os.ReadDir(folder)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		// ignore directories and non executable files
		fileInfo, _ := entry.Info()
		size := fileInfo.Size()
		fileMode := fileInfo.Mode()
		isExecutable := fileMode&0100 != 0
		isFile := !entry.IsDir()
		if isFile && isExecutable && size > 0 {
			serviceInfo, found := ls.services[entry.Name()]
			if !found {
				serviceInfo = &launcher.ServiceInfo{
					Name:    entry.Name(),
					Path:    path.Join(folder, entry.Name()),
					Uptime:  0,
					Running: false,
				}
				ls.services[serviceInfo.Name] = serviceInfo

			}
			serviceInfo.ModifiedTime = fileInfo.ModTime().Format(time.RFC3339)
			serviceInfo.Size = size
		}
	}

	return nil
}

// List the available services and their running status
// This returns the list of services sorted by name
func (ls *LauncherService) List(_ context.Context) ([]launcher.ServiceInfo, error) {
	res := make([]launcher.ServiceInfo, 0)
	ls.mux.Lock()
	defer ls.mux.Unlock()

	// sort the service names
	keys := make([]string, 0, len(ls.services))
	for key := range ls.services {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		svcInfo := ls.services[key]
		ls.updateStatus(svcInfo)
		res = append(res, *svcInfo)
	}
	return res, nil

}

// Start a service
func (ls *LauncherService) Start(
	_ context.Context, name string) (info launcher.ServiceInfo, err error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	logrus.Infof("Starting service '%s'", name)

	serviceInfo, found := ls.services[name]
	if !found {
		info.Error = fmt.Sprintf("service '%s' not found", name)
		logrus.Error(err)
		return info, errors.New(info.Error)
	}
	if serviceInfo.Running {
		err = fmt.Errorf("starting service '%s' failed. The service is already running", name)
		logrus.Error(err)
		return *serviceInfo, err
	}
	svcCmd, hasCmd := ls.cmds[name]
	if hasCmd {
		err = fmt.Errorf("command for service '%s' is exists. This is unexpected. Last error: %s", name, svcCmd.Status().Error)
		logrus.Error(err)
		return *serviceInfo, err
	}

	//svcCmd := exec.Command(serviceInfo.Path)
	// how to pass process stderr to launcher stderr using go-cmd?
	svcCmd = cmd.NewCmd(serviceInfo.Path)

	ls.cmds[name] = svcCmd
	// TODO: should stdout/stderr be routed to the launcher?
	statusChan := svcCmd.Start()
	serviceInfo.StartCount++
	serviceInfo.Running = true

	go func() {
		// cleanup after the process ends
		status := <-statusChan
		_ = status
		ls.mux.Lock()
		defer ls.mux.Unlock()
		// TODO: send event when started and stopped
		ls.updateStatus(serviceInfo)
		logrus.Infof("Service '%s' exited with exit code %d \n%s", name, status.Exit, serviceInfo.Error)
		delete(ls.cmds, name)
	}()

	// FIXME: wait until started
	time.Sleep(time.Millisecond * 10)

	startTime := time.Unix(0, svcCmd.Status().StartTs)
	serviceInfo.StartTime = startTime.Format(time.RFC3339)

	ls.updateStatus(serviceInfo)
	return *serviceInfo, err
}

// Stop a service
func (ls *LauncherService) Stop(_ context.Context, name string) (info launcher.ServiceInfo, err error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	logrus.Infof("Stopping service %s", name)

	serviceInfo, found := ls.services[name]
	if !found {
		info.Error = fmt.Sprintf("service '%s' not found", name)
		logrus.Error(info.Error)
		return info, errors.New(info.Error)
	}
	svcCmd, found := ls.cmds[name]
	if !found {
		err = fmt.Errorf("service '%s' not running", name)
		logrus.Error(err)
		return *serviceInfo, err
	}
	err = svcCmd.Stop()
	// FIXME: wait until stopped?
	time.Sleep(time.Millisecond * 10)
	serviceInfo.Error = "stopped by user"

	// Check that PID is no longer running
	// On Linux FindProcess always succeeds
	proc, _ := os.FindProcess(serviceInfo.PID)
	if proc != nil {
		err = proc.Signal(syscall.Signal(0))
		if err == nil {
			// unexpected
			err = proc.Kill()
			logrus.Errorf("Stop of service '%s' with PID %d failed. Attempt a kill", name, serviceInfo.PID)
		}
	}
	err = nil

	// remove command
	delete(ls.cmds, name)

	return *serviceInfo, err
}

// StopAll stops all running services
func (ls *LauncherService) StopAll(_ context.Context) (err error) {
	logrus.Infof("Stopping all (%d) services", len(ls.cmds))

	ls.mux.Lock()
	defer ls.mux.Unlock()
	// get the list of running services
	names := make([]string, 0)
	for name, _ := range ls.cmds {
		names = append(names, name)
	}
	// stop each service
	for _, name := range names {
		cmd := ls.cmds[name]
		cmd.Stop()
		delete(ls.cmds, name)
	}
	return err
}

// updateStatus updates the service status
func (ls *LauncherService) updateStatus(svcInfo *launcher.ServiceInfo) {
	svcCmd, found := ls.cmds[svcInfo.Name]
	if !found {
		svcInfo.Running = false
		svcInfo.PID = 0
		svcInfo.CPU = 0
		svcInfo.MEM = 0
		return
	}
	svcStatus := svcCmd.Status()
	if svcStatus.Error != nil {
		svcInfo.Error = svcStatus.Error.Error()
	} else if len(svcStatus.Stderr) > 0 {
		lastErr := strings.Join(svcStatus.Stderr[(len(svcStatus.Stderr)-1):], "")
		svcInfo.Error = lastErr
	}
	svcInfo.PID = svcStatus.PID
	svcInfo.Running = (svcStatus.StopTs == 0) && (svcStatus.PID != 0)
	// StartTS is 0 if service hasn't started
	if svcStatus.StartTs == 0 && svcStatus.Error == nil {
		svcInfo.Error = fmt.Sprintf("'%s' failed to start. Exit code %d", svcCmd.Name, svcStatus.Exit)
	}
}

// NewLauncherService returns a new launcher instance for the services in the given folder.
// This scans the folder for executables and adds these to the list of available services
func NewLauncherService(serviceFolder string) *LauncherService {
	logrus.Infof("creating new launcher service with serviceFolder %s", serviceFolder)
	ls := &LauncherService{
		services: make(map[string]*launcher.ServiceInfo),
		cmds:     make(map[string]*cmd.Cmd),
	}
	err := ls.findServices(serviceFolder)
	if err != nil {
		logrus.Error(err)
	}
	return ls
}
