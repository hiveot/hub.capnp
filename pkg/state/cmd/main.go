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

// Connect the service
func main() {
	f := svcconfig.GetFolders("", false)
	// set config defaults
	var cfg = config.NewStateConfig(f.Stores)
	cfg.Backend = bucketstore.BackendKVBTree
	f = svcconfig.LoadServiceConfig(state.ServiceName, false, &cfg)

	svc := statekvstore.NewStateStoreService(cfg)

	listener.RunService(state.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			err = capnpserver.StartStateCapnpServer(svc, lis)
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})
}
