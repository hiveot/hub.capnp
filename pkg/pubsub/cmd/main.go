package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/capnpserver"
	"github.com/hiveot/hub/pkg/pubsub/service"
)

// Start the history store service
func main() {
	f := svcconfig.LoadServiceConfig(pubsub.ServiceName, false, nil)

	svc := service.NewPubSubService()

	listener.RunService(pubsub.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start()
			err = capnpserver.StartPubSubCapnpServer(ctx, lis, svc)
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})

}
