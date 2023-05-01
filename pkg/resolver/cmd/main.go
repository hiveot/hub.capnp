package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpserver"
	"github.com/hiveot/hub/pkg/resolver/service"
)

// main launches the resolver service
func main() {
	//resolverSocketPath := resolver.DefaultResolverPath
	f, _, _ := svcconfig.SetupFolderConfig(resolver.ServiceName)
	svc := service.NewResolverService(f.Run)

	// the resolver uses unix sockets to listen for incoming connections
	listener.RunService(resolver.ServiceName, resolver.DefaultResolverPath, //f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			if err == nil {
				capnpserver.StartResolverServiceCapnpServer(svc, lis, svc.HandleUnknownMethod)
			}
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})
}
