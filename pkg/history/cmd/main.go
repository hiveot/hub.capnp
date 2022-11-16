// Package main with the history store
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/history/service/mongohs"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpserver"
	"github.com/hiveot/hub/pkg/history/config"
	"github.com/hiveot/hub/pkg/launcher"
)

// Start the history store service
func main() {
	logging.SetLogging("info", "")
	ctx := context.Background()

	cfg := config.NewHistoryConfig()
	f := svcconfig.LoadServiceConfig(launcher.ServiceName, false, &cfg)

	srvListener := listener.CreateServiceListener(f.Run, history.ServiceName)

	// For now only mongodb is supported
	svc := mongohs.NewMongoHistoryServer(cfg)
	err := svc.Start(ctx)
	defer svc.Stop(ctx)

	if err == nil {
		logrus.Infof("HistoryServiceCapnpServer starting on: %s", srvListener.Addr())
		_ = capnpserver.StartHistoryCapnpServer(context.Background(), srvListener, svc)
	}
	if err != nil {
		msg := fmt.Sprintf("ERROR: Service '%s' failed to start: %s\n", history.ServiceName, err)
		logrus.Fatal(msg)
	}
	logrus.Infof("History service ended gracefully")
	os.Exit(0)
}
