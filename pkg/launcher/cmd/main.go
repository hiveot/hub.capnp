package main

import (
	"context"
	"flag"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capnpserver"
	"github.com/hiveot/hub/pkg/launcher/service"
)

var binFolder string
var homeFolder string

// Start the launcher service
func main() {
	var err error
	var svc *service.LauncherService
	var ctx = context.Background()
	logging.SetLogging("info", "")

	f := folders.GetFolders("", false)
	// option to override the location of services. Intended for testing
	flag.StringVar(&f.Services, "services", f.Services, "Services folder")
	flag.Parse()

	srvListener := listener.CreateServiceListener(f.Run, launcher.ServiceName)

	if err == nil {
		svc = service.NewLauncherService(f.Services)
	}
	if err == nil {
		err = capnpserver.StartLauncherCapnpServer(ctx, srvListener, svc)
	}
}
