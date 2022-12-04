package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/service"
)

// main launches the gateway service
func main() {
	f := svcconfig.LoadServiceConfig(gateway.ServiceName, false, nil)

	svc := service.NewGatewayService(f.Run)

	listener.RunService(gateway.ServiceName, f.Run,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			//tlsListener := ListenerToTLS(lis)
			err := svc.Start(ctx)
			if err == nil {
				err = capnpserver.StartGatewayServiceCapnpServer(ctx, lis, svc, f.Run)
			}
			return err
		}, func() error {
			// shutdown
			return svc.Stop()
		})
}
