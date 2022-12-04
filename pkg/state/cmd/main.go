package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/state"
	"github.com/hiveot/hub/pkg/state/capnpserver"
	"github.com/hiveot/hub/pkg/state/config"
	statekvstore "github.com/hiveot/hub/pkg/state/service"
)

// Start the service
func main() {
	f := svcconfig.GetFolders("", false)
	// set config defaults
	var config = config.NewStateConfig(f.Stores)
	config.Backend = bucketstore.BackendKVBTree
	f = svcconfig.LoadServiceConfig(state.ServiceName, false, &config)

	svc := statekvstore.NewStateStoreService(config)

	listener.RunService(state.ServiceName, f.Run,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			err = capnpserver.StartStateCapnpServer(ctx, lis, svc)
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})
}
