package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capnpserver"
	"github.com/hiveot/hub/pkg/authn/config"
	"github.com/hiveot/hub/pkg/authn/service"
)

// main entry point to start the authentication service
func main() {
	// get defaults
	f, _, _ := svcconfig.SetupFolderConfig(authn.ServiceName)
	authServiceConfig := config.NewAuthnConfig(f.Stores)
	_ = f.LoadConfig(&authServiceConfig)

	svc := service.NewAuthnService(authServiceConfig)

	listener.RunService(authn.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			if err == nil {
				err = capnpserver.StartAuthnCapnpServer(svc, lis)
			}
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})
}
