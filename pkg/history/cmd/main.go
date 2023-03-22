// Package main with the history store
package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/bucketstore/cmd"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpserver"
	"github.com/hiveot/hub/pkg/history/config"
	"github.com/hiveot/hub/pkg/history/service"
	"github.com/hiveot/hub/pkg/pubsub/capnpclient"
)

// Connect the history store service
func main() {
	var fullUrl = "" // TODO, from config
	ctx := context.Background()
	f, clientCert, caCert := svcconfig.SetupFolderConfig(history.ServiceName)
	cfg := config.NewHistoryConfig(f.Stores)
	_ = f.LoadConfig(&cfg)

	// the service uses the bucket store to store history
	store := cmd.NewBucketStore(cfg.Directory, cfg.ServiceID, cfg.Backend)

	// the service receives the events to store from pubsub.
	conn, err := hubclient.ConnectToService(fullUrl, 0, clientCert, caCert)
	pubSubClient := capnpclient.NewPubSubCapnpClient(ctx, conn)
	svcPubSub, err := pubSubClient.CapServicePubSub(ctx, cfg.ServiceID)
	if err != nil {
		panic("can't connect to pubsub")
	}

	svc := service.NewHistoryService(&cfg, store, svcPubSub)

	listener.RunService(history.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err = svc.Start()
			if err == nil {
				err = capnpserver.StartHistoryServiceCapnpServer(svc, lis)
			}
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			pubSubClient.Release()
			return err
		})

}
