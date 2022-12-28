// Package main with the thing directory store
package main

import (
	"context"
	"net"
	"path/filepath"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpserver"
	"github.com/hiveot/hub/pkg/directory/service"
)

// name of the storage file
const storeFile = "directorystore.json"

// Connect the service
func main() {
	hubID := "urn:hub" // FIXME: get HubID from the Hub somewhere
	f := svcconfig.LoadServiceConfig(directory.ServiceName, false, nil)

	storePath := filepath.Join(f.Stores, directory.ServiceName, storeFile)
	svc := service.NewDirectoryService(hubID, storePath)

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
			return err
		})
}
