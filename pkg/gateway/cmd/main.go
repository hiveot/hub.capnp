package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/service"
)

// main launches the gateway service
func main() {

	logging.SetLogging("info", "")
	ctx := context.Background()

	f := svcconfig.LoadServiceConfig(gateway.ServiceName, false, nil)

	// parse commandline and create server listening socket
	srvListener := listener.CreateServiceListener(f.Run, gateway.ServiceName)

	svc := service.NewGatewayService(f.Run)
	err := svc.Start(ctx)
	defer svc.Stop(ctx)

	if err == nil {
		logrus.Infof("GatewayServiceCapnpServer starting on: %s", srvListener.Addr())
		err = capnpserver.StartGatewayServiceCapnpServer(ctx, srvListener, svc, f.Run)
	}
	if err != nil {
		msg := fmt.Sprintf("ERROR: Gateway service failed to start: %s\n", err)
		logrus.Fatal(msg)
	}
	logrus.Infof("Gatewy service ended gracefully")
	os.Exit(0)
}
