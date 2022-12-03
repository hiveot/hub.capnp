package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capnpserver"
	"github.com/hiveot/hub/pkg/authn/config"
	"github.com/hiveot/hub/pkg/authn/service"
)

const DefaultUserConfigFolderName = "configStore"

func Main() {
	main()
}

// main entry point to start the authentication service
func main() {
	// get defaults
	f := svcconfig.GetFolders("", false)
	authServiceConfig := config.NewAuthnConfig(f.Stores)
	f = svcconfig.LoadServiceConfig(authn.ServiceName, false, &authServiceConfig)

	// parse commandline and create server listening socket
	srvListener := listener.CreateServiceListener(f.Run, authn.ServiceName)
	ctx := context.Background()

	svc := service.NewAuthnService(ctx, authServiceConfig)
	err := svc.Start(ctx)
	if err == nil {
		defer svc.Stop(ctx)
	}
	if err == nil {
		logrus.Infof("AuthnCapnpServer starting on: %s", srvListener.Addr())
		err = capnpserver.StartAuthnCapnpServer(ctx, srvListener, svc)
	}
	if err != nil {
		msg := fmt.Sprintf("ERROR: Service '%s' failed to start: %s\n", authn.ServiceName, err)
		logrus.Fatal(msg)
	}
	logrus.Warningf("Authn service ended gracefully")

	os.Exit(0)
}
