package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capnpserver"
	"github.com/hiveot/hub/pkg/launcher/config"
	"github.com/hiveot/hub/pkg/launcher/service"
)

// Connect the launcher service
func main() {
	var lc config.LauncherConfig

	lc = config.NewLauncherConfig()
	f := svcconfig.LoadServiceConfig(launcher.ServiceName, false, &lc)

	svc := service.NewLauncherService(f, lc)

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
