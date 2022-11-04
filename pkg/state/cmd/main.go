package main

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/bucketstore"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/state"
	"github.com/hiveot/hub/pkg/state/capnpserver"
	"github.com/hiveot/hub/pkg/state/config"
	statekvstore "github.com/hiveot/hub/pkg/state/service"
)

// Start the launcher service
func main() {
	var err error
	var svc state.IStateStore
	var ctx = context.Background()

	logrus.SetLevel(logrus.InfoLevel)
	// this is a service so go 2 levels up
	f := svcconfig.GetFolders("", false)
	var stateConfig = config.NewStateConfig(f.Stores)
	stateConfig.Backend = bucketstore.BackendKVStore
	//stateConfig.Backend = bucketstore.BackendBBolt
	//stateConfig.Backend = bucketstore.BackendPebble

	// option to override the location of the store itself. Intended for testing
	//flag.StringVar(&stateConfig.DatabaseURL, "DB", stateConfig.DatabaseURL, "State store file")
	f = svcconfig.LoadServiceConfig(launcher.ServiceName, false, &stateConfig)

	srvListener := listener.CreateServiceListener(f.Run, certs.ServiceName)

	if err == nil {
		svc = statekvstore.NewStateStoreService(stateConfig)
	}
	if err == nil {
		err = capnpserver.StartStateCapnpServer(ctx, srvListener, svc)
	}
}
