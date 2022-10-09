package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/struCoder/pidusage"

	"github.com/hiveot/hub/pkg/launcher"
)

// LauncherService manages starting and stopping of services
// This implements the ILauncher interface
type LauncherService struct {
	// map of service name to running status
	services map[string]*launcher.ServiceInfo
	// map of started commands
	cmds map[string]*exec.Cmd
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
		err = fmt.Errorf("process for service '%s' already exists using PID %d. This is unexpected.", name, svcCmd.Process.Pid)
		logrus.Error(err)
		return *serviceInfo, err
	}

	//svcCmd := exec.Command(serviceInfo.Path)
	// how to pass process stderr to launcher stderr using go-cmd?
	svcCmd = exec.Command(serviceInfo.Path)

	err = svcCmd.Start()
	if err != nil {
		serviceInfo.Error = fmt.Sprintf("failed starting '%s': %s", name, err.Error())
		logrus.Error(serviceInfo.Error)
		err = errors.New(serviceInfo.Error)
		return *serviceInfo, err
	}
	ls.cmds[name] = svcCmd
	logrus.Infof("Service '%s' has started", name)

	// TODO: should stdout/stderr be routed to the launcher?
	serviceInfo.StartTime = time.Now().Format(time.RFC3339)
	serviceInfo.PID = svcCmd.Process.Pid
	serviceInfo.Error = ""
	serviceInfo.StartCount++
	serviceInfo.Running = true

	go func() {
		// cleanup after the process ends
		status := svcCmd.Wait()
		_ = status
		ls.mux.Lock()
		defer ls.mux.Unlock()

		// TODO: send event when started and stopped
		serviceInfo.StopTime = time.Now().Format(time.RFC3339)
		serviceInfo.Running = false
		// processState holds exit info
		procState := svcCmd.ProcessState
		if status != nil {
			serviceInfo.Error = fmt.Sprintf("Service '%s' has stopped with: %s", name, status.Error())
		} else if procState != nil {
			serviceInfo.Error = fmt.Sprintf("Service '%s' has stopped with exit code %d: sys='%v'", name, procState.ExitCode(), procState.Sys())
		} else {
			serviceInfo.Error = fmt.Sprintf("Service '%s' has stopped without info", name)
		}
		logrus.Infof(serviceInfo.Error)
		ls.updateStatus(serviceInfo)
		delete(ls.cmds, name)
	}()

	// FIXME: wait until started
	time.Sleep(time.Millisecond * 10)

	startTime := time.Now()
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
	err = svcCmd.Process.Signal(syscall.SIGTERM)
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
			errmsg := fmt.Sprintf("Stop of service '%s' with PID %d failed. Attempt a kill", name, serviceInfo.PID)
			serviceInfo.Error = errmsg
			logrus.Error(errmsg)
		}
	}
	err = nil

	// remove command
	delete(ls.cmds, name)

	return *serviceInfo, err
}

// StopAll stops all running services
func (ls *LauncherService) StopAll(ctx context.Context) (err error) {
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
		ls.Stop(ctx, name)
		delete(ls.cmds, name)
	}
	return err
}

// updateStatus updates the service  status
func (ls *LauncherService) updateStatus(svcInfo *launcher.ServiceInfo) {
	pidStats, _ := pidusage.GetStat(svcInfo.PID)
	if pidStats != nil {
		svcInfo.RSS = int(pidStats.Memory) // RSS is in KB
		svcInfo.CPU = int(pidStats.CPU)
	} else {
		svcInfo.CPU = 0
		svcInfo.RSS = 0
	}
}

// NewLauncherService returns a new launcher instance for the services in the given folder.
// This scans the folder for executables and adds these to the list of available services
func NewLauncherService(serviceFolder string) *LauncherService {
	logrus.Infof("creating new launcher service with serviceFolder %s", serviceFolder)
	ls := &LauncherService{
		services: make(map[string]*launcher.ServiceInfo),
		cmds:     make(map[string]*exec.Cmd),
	}
	err := ls.findServices(serviceFolder)
	if err != nil {
		logrus.Error(err)
	}
	return ls
}
