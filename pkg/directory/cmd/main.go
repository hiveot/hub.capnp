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
	"github.com/hiveot/hub/pkg/directory/service/directorykvstore"
	"github.com/hiveot/hub/pkg/launcher"
)

// name of the storage file
const storeFile = "directorystore.json"

// Use the commandline option -f path/to/store.json for the storage file
func main() {
	logging.SetLogging("info", "")

	f := svcconfig.LoadServiceConfig(launcher.ServiceName, false, nil)
	storePath := filepath.Join(f.Stores, directory.ServiceName, storeFile)

	// parse commandline and create server listening socket
	srvListener := listener.CreateServiceListener(f.Run, directory.ServiceName)
	ctx := context.Background()

	svc, err := directorykvstore.NewDirectoryKVStoreServer(ctx, storePath)
	if err == nil {
		err = svc.Start(ctx)
		defer svc.Stop()
	}

	if err == nil {
		logrus.Infof("DirectoryCapnpServer starting on: %s", srvListener.Addr())
		err = capnpserver.StartDirectoryCapnpServer(ctx, srvListener, svc)
	}
	if err != nil {
		msg := fmt.Sprintf("ERROR: Service '%s' failed to start: %s\n", directory.ServiceName, err)
		logrus.Fatal(msg)
	}
	logrus.Warningf("Directory ended gracefully")

	os.Exit(0)
}
