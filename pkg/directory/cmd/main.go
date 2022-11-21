// Package main with the thing directory store
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpserver"
	"github.com/hiveot/hub/pkg/directory/service"
	"github.com/hiveot/hub/pkg/launcher"
)

// name of the storage file
const storeFile = "directorystore.json"

// Use the commandline option -f path/to/store.json for the storage file
func main() {
	logging.SetLogging("info", "")
	ctx := context.Background()
	hubID := "urn:hub" // FIXME: get HubID from the Hub somewhere

	f := svcconfig.LoadServiceConfig(launcher.ServiceName, false, nil)
	storePath := filepath.Join(f.Stores, directory.ServiceName, storeFile)

	// parse commandline and create server listening socket
	srvListener := listener.CreateServiceListener(f.Run, directory.ServiceName)

	svc := service.NewDirectoryService(ctx, hubID, storePath)
	err := svc.Start(ctx)
	defer svc.Stop(ctx)

	if err == nil {
		logrus.Infof("DirectoryCapnpServer starting on: %s", srvListener.Addr())
		err = capnpserver.StartDirectoryCapnpServer(ctx, srvListener, svc)
	}
	if err != nil {
		msg := fmt.Sprintf("ERROR: Service '%s' failed to start: %s\n", directory.ServiceName, err)
		logrus.Fatal(msg)
	}
	logrus.Infof("Directory service ended gracefully")
	os.Exit(0)
}
