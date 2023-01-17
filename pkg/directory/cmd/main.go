// Package main with the thing directory store
package main

import (
	"context"
	"net"
	"path/filepath"

	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/bucketstore/kvbtree"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpserver"
	"github.com/hiveot/hub/pkg/directory/service"
	"github.com/hiveot/hub/pkg/pubsub/capnpclient"
)

// name of the storage file
const storeFile = "directorystore.json"

// Connect the service
func main() {
	ctx := context.Background()
	serviceID := directory.ServiceName
	f, clientCert, caCert := svcconfig.SetupFolderConfig(directory.ServiceName)

	// the service uses the bucket store to store directory entries
	storePath := filepath.Join(f.Stores, directory.ServiceName, storeFile)
	store := kvbtree.NewKVStore(directory.ServiceName, storePath)

	// the service stores captured TD events from pubsub
	conn, err := hubclient.ConnectToHub("", "", clientCert, caCert)
	pubSubClient := capnpclient.NewPubSubCapnpClient(ctx, conn)
	svcPubSub, err := pubSubClient.CapServicePubSub(ctx, serviceID)
	if err != nil {
		panic("can't connect to pubsub")
	}

	svc := service.NewDirectoryService(serviceID, store, svcPubSub)

	listener.RunService(directory.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			if err == nil {
				err = capnpserver.StartDirectoryServiceCapnpServer(svc, lis)
			}
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			pubSubClient.Release()
			return err
		})
}
