package main

import (
	"context"
	"net"
	"os"
	"path/filepath"

	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/capnpserver"
	"github.com/hiveot/hub/pkg/authz/service"
)

const aclStoreFile = "authz.acl"

// main entry point to start the authorization service
func main() {
	f, _, _ := svcconfig.SetupFolderConfig(authz.ServiceName)
	aclStoreFolder := filepath.Join(f.Stores, authz.ServiceName)
	aclStorePath := filepath.Join(aclStoreFolder, aclStoreFile)
	_ = os.Mkdir(aclStoreFolder, 0700)

	svc := service.NewAuthzService(aclStorePath)

	listener.RunService(authz.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			err = capnpserver.StartAuthzCapnpServer(svc, lis)
			return err
		}, func() error {
			// shutdown
			svc.Stop()
			return nil
		})
}
