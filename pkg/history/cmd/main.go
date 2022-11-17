// Package main with the history store
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/bucketstore/cmd"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpserver"
	"github.com/hiveot/hub/pkg/history/config"
	"github.com/hiveot/hub/pkg/history/service"
	"github.com/hiveot/hub/pkg/launcher"
)

// Start the history store service
func main() {
	logging.SetLogging("info", "")
	ctx := context.Background()

	f := svcconfig.GetFolders("", false)
	cfg := config.NewHistoryConfig(f.Stores)
	f = svcconfig.LoadServiceConfig(launcher.ServiceName, false, &cfg)

	srvListener := listener.CreateServiceListener(f.Run, history.ServiceName)

	// the service uses the bucket store
	store := cmd.NewBucketStore(cfg.Directory, history.ServiceName, cfg.Backend)
	err := store.Open()
	defer store.Close()

	svc := service.NewHistoryService(store)
	err = svc.Start(ctx)
	if err != nil {
		logrus.Panicf("unable launch the history service: %s", err)
	}

	// connections go via capnp RPC
	if err == nil {
		logrus.Infof("HistoryServiceCapnpServer starting on: %s", srvListener.Addr())
		_ = capnpserver.StartHistoryServiceCapnpServer(context.Background(), srvListener, svc)
	}
	if err != nil {
		msg := fmt.Sprintf("ERROR: Service '%s' failed to start: %s\n", history.ServiceName, err)
		logrus.Fatal(msg)
	}
	logrus.Infof("History service ended gracefully")
	err = svc.Stop(ctx)
	os.Exit(0)
}
