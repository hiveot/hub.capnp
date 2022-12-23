package main

import (
	"context"
	"net"
	"os"
	"path/filepath"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/capnpserver"
	"github.com/hiveot/hub/pkg/authz/service"
)

const aclStoreFile = "authz.acl"

// main entry point to start the authorization service
func main() {
	f := svcconfig.LoadServiceConfig(authz.ServiceName, false, nil)
	aclStoreFolder := filepath.Join(f.Stores, authz.ServiceName)
	aclStorePath := filepath.Join(aclStoreFolder, aclStoreFile)
	_ = os.Mkdir(aclStoreFolder, 0700)

	svc := service.NewAuthzService(aclStorePath)

	listener.RunService(authz.ServiceName, f.Run,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			err = capnpserver.StartAuthzCapnpServer(lis, svc)
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})
}
