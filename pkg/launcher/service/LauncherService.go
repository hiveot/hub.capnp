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
	"sync/atomic"
	"time"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/struCoder/pidusage"
	"gopkg.in/fsnotify.v1"

	"github.com/hiveot/hub/lib/svcconfig"
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
	// list of started commands in startup order
	cmds []*exec.Cmd

	// mutex to keep things safe
	mux sync.Mutex
	// watch service and binding folders for updates
	serviceWatcher *fsnotify.Watcher
	// service is running
	isRunning atomic.Bool
}

// Add newly discovered executable services
// If the service is already know, only update its size and timestamp
func (svc *LauncherService) findServices(folder string) error {
	count := 0
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
			count++
			serviceInfo, found := svc.services[entry.Name()]
			if !found {
				serviceInfo = &launcher.ServiceInfo{
					Name:    entry.Name(),
					Path:    path.Join(folder, entry.Name()),
					Uptime:  0,
					Running: false,
				}
				svc.services[serviceInfo.Name] = serviceInfo
			}
			serviceInfo.ModifiedTime = fileInfo.ModTime().Format(time.RFC3339)
			serviceInfo.Size = size
		}
	}
	logrus.Infof("found '%d' services in '%s'", count, folder)
	return nil
}

// List all available or just the running services and their status
// This returns the list of services sorted by name
func (svc *LauncherService) List(_ context.Context, onlyRunning bool) ([]launcher.ServiceInfo, error) {
	svc.mux.Lock()
	defer svc.mux.Unlock()

	// get the keys of the services to include and sort them
	keys := make([]string, 0, len(svc.services))
	for key, val := range svc.services {
		if !onlyRunning || val.Running {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	res := make([]launcher.ServiceInfo, 0, len(keys))
	for _, key := range keys {
		svcInfo := svc.services[key]
		svc.updateStatus(svcInfo)
		res = append(res, *svcInfo)
	}
	return res, nil
}

// ScanServices scans the service and bindings folders for changes and updates the
// services list
func (svc *LauncherService) ScanServices(ctx context.Context) error {
	err := svc.findServices(svc.f.Services)
	if err != nil {
		logrus.Error(err)
		return err
	}
	// ignore bindings folder if not set
	if svc.f.Bindings != "" {
		err = svc.findServices(svc.f.Bindings)
		if err != nil {
			logrus.Error(err)
			return err
		}
	}
	return nil
}
func (svc *LauncherService) StartService(
	_ context.Context, name string) (info launcher.ServiceInfo, err error) {
	svc.mux.Lock()
	defer svc.mux.Unlock()

	// step 1: pre-checks
	serviceInfo, found := svc.services[name]
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
	// don't start twice
	for _, cmd := range svc.cmds {
		if cmd.Path == serviceInfo.Path {
			err = fmt.Errorf("process for service '%s' already exists using PID %d",
				serviceInfo.Name, cmd.Process.Pid)
			logrus.Error(err)
			return *serviceInfo, err
		}
	}

	// step 2: create the command to start the service ... but wait for step 3
	svcCmd := exec.Command(serviceInfo.Path)

	// step3: setup logging before starting service
	logrus.Infof("Starting service '%s'", name)

	if svc.cfg.LogServices {
		// inspired by https://gist.github.com/jerblack/4b98ba48ed3fb1d9f7544d2b1a1be287
		logfile := path.Join(svc.f.Logs, name+".log")
		logrus.Debugf("creating new logfile %s", logfile)
		fp, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err == nil {
			if svc.cfg.AttachStderr {
				// log stderr to launcher stderr and to file
				multiwriter := io.MultiWriter(os.Stderr, fp)
				svcCmd.Stderr = multiwriter
			} else {
				// just log stderr to file
				svcCmd.Stderr = fp
			}
			if svc.cfg.AttachStdout {
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
		if svc.cfg.AttachStderr {
			svcCmd.Stderr = os.Stderr
		}
		if svc.cfg.AttachStdout {
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
	svc.cmds = append(svc.cmds, svcCmd)
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
		svc.mux.Lock()
		defer svc.mux.Unlock()

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
		svc.updateStatus(serviceInfo)
		// find the service to delete
		i := lo.IndexOf(svc.cmds, svcCmd)
		//lo.Delete(svc.cmds, i)  - why doesn't this exist?
		svc.cmds = append(svc.cmds[:i], svc.cmds[i+1:]...) // this is so daft!
	}()

	// Give it some time to get up and running in case it is needed as a dependency
	// TODO: wait for channel
	time.Sleep(time.Millisecond * 100)

	// last, update the CPU and memory status
	svc.updateStatus(serviceInfo)
	return *serviceInfo, err
}

// StartAll starts all enabled services
func (svc *LauncherService) StartAll(ctx context.Context) (err error) {
	logrus.Infof("Starting all enabled services")

	// ensure they start in order
	for _, svcName := range svc.cfg.Autostart {
		svcInfo := svc.services[svcName]
		if svcInfo != nil && svcInfo.Running {
			// skip
		} else {
			_, err2 := svc.StartService(ctx, svcName)
			if err2 != nil {
				err = err2
			}
		}
	}
	// start the remaining services
	for svcName, svcInfo := range svc.services {
		if !svcInfo.Running {
			_, err2 := svc.StartService(ctx, svcName)
			if err2 != nil {
				err = err2
			}
		}
	}
	return err
}

func (svc *LauncherService) StopService(_ context.Context, name string) (info launcher.ServiceInfo, err error) {
	logrus.Infof("Stopping service %s", name)

	serviceInfo, found := svc.services[name]
	if !found {
		info.Status = fmt.Sprintf("service '%s' not found", name)
		err = errors.New(info.Status)
		logrus.Error(err)
		return info, err
	}
	err = Stop(serviceInfo.Name, serviceInfo.PID)
	if err == nil {
		svc.mux.Lock()
		defer svc.mux.Unlock()
		serviceInfo.Running = false
		serviceInfo.Status = "stopped by user"
	}
	return *serviceInfo, err
}

// StopAll stops all running services in reverse order they were started
func (svc *LauncherService) StopAll(ctx context.Context) (err error) {

	svc.mux.Lock()
	logrus.Infof("Stopping all (%d) services", len(svc.cmds))

	// use a copy of the commands as the command list will be mutated
	cmdsToStop := svc.cmds[:]

	svc.mux.Unlock()

	// stop each service
	for i := len(cmdsToStop) - 1; i >= 0; i-- {
		c := cmdsToStop[i]
		err = Stop(c.Path, c.Process.Pid)
	}
	time.Sleep(time.Millisecond)
	return err
}

// updateStatus updates the service  status
func (svc *LauncherService) updateStatus(svcInfo *launcher.ServiceInfo) {
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

// WatchServices watches the services and bindings folder for changes and reloads
// This will detect adding new services or bindings without requiring a restart.
func (svc *LauncherService) WatchServices(ctx context.Context) error {
	svc.serviceWatcher, _ = fsnotify.NewWatcher()
	err := svc.serviceWatcher.Add(svc.f.Services)
	if err == nil && svc.f.Bindings != "" {
		err = svc.serviceWatcher.Add(svc.f.Bindings)
	}
	if err == nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					logrus.Infof("service watcher ended by context")
					return
				case event := <-svc.serviceWatcher.Events:
					isRunning := svc.isRunning.Load()
					if isRunning {
						logrus.Infof("watcher event: %v", event)
						_ = svc.ScanServices(ctx)
					} else {
						logrus.Infof("service watcher stopped")
						return
					}
				case err := <-svc.serviceWatcher.Errors:
					logrus.Errorf("error: %s", err)
				}
			}
		}()

	}
	return err
}

// Start the launcher service
func (svc *LauncherService) Start(ctx context.Context) error {
	svc.isRunning.Store(true)

	_ = svc.WatchServices(ctx)
	err := svc.ScanServices(ctx)
	if err != nil {
		return err
	}

	// autostart the services
	for _, name := range svc.cfg.Autostart {
		_, err2 := svc.StartService(ctx, name)
		if err2 != nil {
			err = err2
		}
	}
	return err
}

// Stop the launcher and all running services
func (svc *LauncherService) Stop() error {
	svc.isRunning.Store(false)
	return svc.StopAll(context.Background())
}

// NewLauncherService returns a new launcher instance for the services in the given services folder.
// This scans the folder for executables, adds these to the list of available services and autostarts services
// Logging will be enabled based on LauncherConfig.
func NewLauncherService(f svcconfig.AppFolders, cfg config.LauncherConfig) *LauncherService {

	ls := &LauncherService{
		f:        f,
		cfg:      cfg,
		services: make(map[string]*launcher.ServiceInfo),
		cmds:     make([]*exec.Cmd, 0),
	}

	return ls
}
