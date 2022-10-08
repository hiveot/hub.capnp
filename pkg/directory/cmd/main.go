// Package main with the thing directory store
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpserver"
	"github.com/hiveot/hub/pkg/directory/service/directorykvstore"
)

// name of the storage file
const storeFile = "directorystore.json"

// Use the commandline option -f path/to/store.json for the storage file
func main() {
	homeFolder := filepath.Join(filepath.Dir(os.Args[0]), "../..")
	f := folders.GetFolders(homeFolder, false)
	storePath := filepath.Join(f.Stores, directory.ServiceName, storeFile)
	flag.StringVar(&storePath, "f", storePath, "File path of the directory store.")
	flag.Parse()

	srvListener := listener.CreateServiceListener(f.Run, directory.ServiceName)

	svc, err := directorykvstore.NewDirectoryKVStoreServer(storePath)
	if err != nil {
		log.Fatalf("Service '%s' failed to start: %s", directory.ServiceName, err)
	}
	logrus.Infof("DirectoryCapnpServer starting on: %s", srvListener.Addr())
	capnpserver.StartDirectoryCapnpServer(context.Background(), srvListener, svc)
}
