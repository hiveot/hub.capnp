package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capnpserver"
	"github.com/hiveot/hub/pkg/authn/config"
	"github.com/hiveot/hub/pkg/authn/service"
)

// main entry point to start the authentication service
func main() {
	// get defaults
	f := svcconfig.GetFolders("", false)
	authServiceConfig := config.NewAuthnConfig(f.Stores)
	f = svcconfig.LoadServiceConfig(authn.ServiceName, false, &authServiceConfig)

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
