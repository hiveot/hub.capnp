package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/struCoder/pidusage"

	"github.com/hiveot/hub.go/pkg/proc"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/config"
)

// LauncherService manages starting and stopping of services
// This implements the ILauncher interface
type LauncherService struct {
	// service configuration
	cfg config.LauncherConfig
	f   svcconfig.AppFolders

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

// List all available or just the running services and their status
// This returns the list of services sorted by name
func (ls *LauncherService) List(_ context.Context, onlyRunning bool) ([]launcher.ServiceInfo, error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	// get the keys of the services to include and sort them
	keys := make([]string, 0, len(ls.services))
	for key, val := range ls.services {
		if !onlyRunning || val.Running {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	res := make([]launcher.ServiceInfo, 0, len(keys))
	for _, key := range keys {
		svcInfo := ls.services[key]
		ls.updateStatus(svcInfo)
		res = append(res, *svcInfo)
	}
	return res, nil

}

func (ls *LauncherService) StartService(
	_ context.Context, name string) (info launcher.ServiceInfo, err error) {
	ls.mux.Lock()
	defer ls.mux.Unlock()

	// step 1: pre-checks
	serviceInfo, found := ls.services[name]
	if !found {
		info.Status = fmt.Sprintf("service '%s' not found", name)
		logrus.Error(info.Status)
		return info, errors.New(info.Status)
	}
	if serviceInfo.Running {
		err = fmt.Errorf("starting service '%s' failed. The service is already running", name)
		logrus.Error(err)
		return *serviceInfo, err
	}
	svcCmd, hasCmd := ls.cmds[name]
	if hasCmd {
		err = fmt.Errorf("process for service '%s' already exists using PID %d. This is unexpected", name, svcCmd.Process.Pid)
		logrus.Error(err)
		return *serviceInfo, err
	}

	// step 2: create the command to start the service ... but wait for step 3
	svcCmd = exec.Command(serviceInfo.Path)

	// step3: setup logging before starting service
	logrus.Infof("Starting service '%s'", name)

	if ls.cfg.LogServices {
		// inspired by https://gist.github.com/jerblack/4b98ba48ed3fb1d9f7544d2b1a1be287
		logfile := path.Join(ls.f.Logs, name+".log")
		logrus.Infof("creating new logfile %s", logfile)
		fp, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err == nil {
			if ls.cfg.AttachStderr {
				// log stderr to launcher stderr and to file
				multiwriter := io.MultiWriter(os.Stderr, fp)
				svcCmd.Stderr = multiwriter
			} else {
				// just log stderr to file
				svcCmd.Stderr = fp
			}
			if ls.cfg.AttachStdout {
				// log stdout to launcher stdout and to file
				multiwriter := io.MultiWriter(os.Stdout, fp)
				svcCmd.Stdout = multiwriter
			} else {
				// just log stdout to file
				svcCmd.Stdout = fp
			}
		} else {
			logrus.Warningf("creating logfile %s failed: %s", logfile, err)
		}
	} else {
		if ls.cfg.AttachStderr {
			svcCmd.Stderr = os.Stderr
		}
		if ls.cfg.AttachStdout {
			svcCmd.Stdout = os.Stdout
		}
	}
	// step 4: start the command and setup serviceInfo
	err = svcCmd.Start()
	if err != nil {
		serviceInfo.Status = fmt.Sprintf("failed starting '%s': %s", name, err.Error())
		err = errors.New(serviceInfo.Status)
		logrus.Error(err)
		return *serviceInfo, err
	}
	ls.cmds[name] = svcCmd
	//logrus.Warningf("Service '%s' has started", name)

	serviceInfo.StartTime = time.Now().Format(time.RFC3339)
	serviceInfo.PID = svcCmd.Process.Pid
	serviceInfo.Status = ""
	serviceInfo.StartCount++
	serviceInfo.Running = true

	// step 5: handle command termination and cleanup
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
			serviceInfo.Status = fmt.Sprintf("Service '%s' has stopped with: %s", name, status.Error())
		} else if procState != nil {
			serviceInfo.Status = fmt.Sprintf("Service '%s' has stopped with exit code %d: sys='%v'", name, procState.ExitCode(), procState.Sys())
		} else {
			serviceInfo.Status = fmt.Sprintf("Service '%s' has stopped without info", name)
		}
		logrus.Warningf(serviceInfo.Status)
		ls.updateStatus(serviceInfo)
		delete(ls.cmds, name)
	}()

	// FIXME: wait until started
	time.Sleep(time.Millisecond * 10)

	// last, update the CPU and memory status
	ls.updateStatus(serviceInfo)
	return *serviceInfo, err
}

// StartAll starts all enabled services
func (ls *LauncherService) StartAll(ctx context.Context) (err error) {
	logrus.Infof("Starting all enabled services")

	for svcName, svcInfo := range ls.services {
		if !svcInfo.Running {
			_, err2 := ls.StartService(ctx, svcName)
			if err2 != nil {
				err = err2
			}
		}
	}
	return err
}

func (ls *LauncherService) StopService(_ context.Context, name string) (info launcher.ServiceInfo, err error) {
	logrus.Infof("Stopping service %s", name)

	serviceInfo, found := ls.services[name]
	if !found {
		info.Status = fmt.Sprintf("service '%s' not found", name)
		err = errors.New(info.Status)
		logrus.Error(err)
		return info, err
	}
	err = proc.Stop(serviceInfo.Name, serviceInfo.PID)
	if err == nil {
		ls.mux.Lock()
		defer ls.mux.Unlock()
		serviceInfo.Running = false
		serviceInfo.Status = "stopped by user"
	}
	return *serviceInfo, err
}

// StopAll stops all running services
func (ls *LauncherService) StopAll(ctx context.Context) (err error) {

	ls.mux.Lock()
	logrus.Infof("Stopping all (%d) services", len(ls.cmds))
	// get the list of running services
	names := make([]string, 0)
	for name := range ls.cmds {
		names = append(names, name)
	}
	ls.mux.Unlock()

	// stop each service
	for _, name := range names {
		_, _ = ls.StopService(ctx, name)
		delete(ls.cmds, name)
	}
	time.Sleep(time.Millisecond)
	return err
}

// updateStatus updates the service  status
func (ls *LauncherService) updateStatus(svcInfo *launcher.ServiceInfo) {
	if svcInfo.PID != 0 {

		//Option A: use pidusage - doesn't work on Windows though
		//warning, pidusage is not very fast
		pidStats, _ := pidusage.GetStat(svcInfo.PID)
		if pidStats != nil {
			svcInfo.RSS = int(pidStats.Memory) // RSS is in KB
			svcInfo.CPU = int(pidStats.CPU)
		} else {
			svcInfo.CPU = 0
			svcInfo.RSS = 0
		}

		// Option B: use go-osstat - slower
		//cpuStat, err := cpu.Get()
		//if err == nil {
		//	svcInfo.CPU = cpuStat.CPUCount // FIXME: this is a counter, not %
		//}
		//memStat, err := memory.Get()
		//if err == nil {
		//	svcInfo.RSS = int(memStat.Used)
		//}

		//Option C: read statm directly. Fastest but only gets memory.
		//path := fmt.Sprintf("/proc/%d/statm", svcInfo.PID)
		//statm, err := ioutil.ReadFile(path)
		//if err == nil {
		//	fields := strings.Split(string(statm), " ")
		//	if len(fields) < 2 {
		//		// invalid data
		//	} else {
		//		rss, err := strconv.ParseInt(fields[1], 10, 64)
		//		if err != nil {
		//			// invalid data
		//		} else {
		//			svcInfo.RSS = int(rss * int64(os.Getpagesize()))
		//		}
		//	}
		//}
	}

}

// Start the launcher service
func (ls *LauncherService) Start(ctx context.Context) error {
	err := ls.findServices(ls.f.Services)
	if err != nil {
		logrus.Error(err)
		return err
	}

	// autostart the services
	for _, name := range ls.cfg.Autostart {
		_, err2 := ls.StartService(ctx, name)
		if err2 != nil {
			err = err2
		}
	}
	return err
}

// Stop the launcher and all running services
func (ls *LauncherService) Stop() error {
	return ls.StopAll(context.Background())
}

// NewLauncherService returns a new launcher instance for the services in the given services folder.
// This scans the folder for executables, adds these to the list of available services and autostarts services
// Logging will be enabled based on LauncherConfig.
func NewLauncherService(f svcconfig.AppFolders, cfg config.LauncherConfig) *LauncherService {

	ls := &LauncherService{
		f:        f,
		cfg:      cfg,
		services: make(map[string]*launcher.ServiceInfo),
		cmds:     make(map[string]*exec.Cmd),
	}

	return ls
}
