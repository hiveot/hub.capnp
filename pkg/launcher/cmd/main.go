package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capnpserver"
	"github.com/hiveot/hub/pkg/launcher/config"
	"github.com/hiveot/hub/pkg/launcher/service"
)

// Connect the launcher service
func main() {
	f, _, _ := svcconfig.SetupFolderConfig(launcher.ServiceName)
	cfg := config.NewLauncherConfig()
	_ = f.LoadConfig(&cfg)

	svc := service.NewLauncherService(f, cfg)

	listener.RunService(launcher.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			if err == nil {
				err = capnpserver.StartLauncherCapnpServer(lis, svc)
			}
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})
}
