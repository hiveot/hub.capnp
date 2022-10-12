package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capnpserver"
	"github.com/hiveot/hub/pkg/launcher/config"
	"github.com/hiveot/hub/pkg/launcher/service"
)

// Start the launcher service
func main() {
	var err error
	var svc *service.LauncherService
	var lc config.LauncherConfig
	var ctx = context.Background()

	f := folders.GetFolders("", false)
	lc = config.NewLauncherConfig()
	f, err = svcconfig.LoadServiceConfig(f, launcher.ServiceName, false, &lc)

	// exit if config is invalid
	if err != nil {
		logrus.Errorf("ERROR: invalid yaml configuration: " + err.Error() + "\n")
		os.Exit(-1)
	}

	srvListener := listener.CreateServiceListener(f.Run, launcher.ServiceName)

	if err == nil {
		svc, err = service.NewLauncherService(ctx, f, lc)
	}
	if err == nil {
		err = capnpserver.StartLauncherCapnpServer(ctx, srvListener, svc)
	}
	if err != nil {
		logrus.Errorf("Launcher startup failed:" + err.Error() + "\n")
		os.Exit(-1)
	}
	logrus.Warningf("Launcher says bye bye :)\n")
}
