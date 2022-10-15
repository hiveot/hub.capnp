// Package main with the history store
package main

import (
	"context"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpserver"
	"github.com/hiveot/hub/pkg/history/config"
	"github.com/hiveot/hub/pkg/history/service/mongohs"
	"github.com/hiveot/hub/pkg/launcher"
)

// Start the history store service
func main() {
	logging.SetLogging("info", "")
	cfg := config.NewHistoryConfig()
	f := svcconfig.LoadServiceConfig(launcher.ServiceName, false, &cfg)

	srvListener := listener.CreateServiceListener(f.Run, history.ServiceName)

	// For now only mongodb is supported
	svc := mongohs.NewMongoHistoryServer(cfg)
	svc.Start()
	defer svc.Stop()

	_ = capnpserver.StartHistoryCapnpServer(context.Background(), srvListener, svc)

}
