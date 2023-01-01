// Package main with the history store
package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/bucketstore/cmd"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpserver"
	"github.com/hiveot/hub/pkg/history/config"
	"github.com/hiveot/hub/pkg/history/service"
)

// Connect the history store service
func main() {

	f := svcconfig.GetFolders("", false)
	cfg := config.NewHistoryConfig(f.Stores)
	f = svcconfig.LoadServiceConfig(history.ServiceName, false, &cfg)

	// the service uses the bucket store
	store := cmd.NewBucketStore(cfg.Directory, cfg.ServiceID, cfg.Backend)
	svc := service.NewHistoryService(store, "urn:"+cfg.ServiceID)

	listener.RunService(history.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := store.Open()
			if err == nil {
				err = svc.Start(ctx)
			}
			if err == nil {
				err = capnpserver.StartHistoryServiceCapnpServer(svc, lis)
			}
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			err2 := store.Close()
			if err == nil {
				err = err2
			}
			return err
		})

}
