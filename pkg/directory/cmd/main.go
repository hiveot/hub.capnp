// Package main with the thing directory store
package main

import (
	"context"
	"flag"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/directory/capnpserver"
	"github.com/hiveot/hub/pkg/directory/service/directorykvstore"
)

// ServiceName is the name of the store for logging
const ServiceName = "directorystore"

// DirectoryStorePath is the path to the storage file for the in-memory store.
const DirectoryStorePath = "config/directorystore.json"

// Use the commandline option -f path/to/store.json for the storage file
func main() {
	storePath := DirectoryStorePath
	flag.StringVar(&storePath, "f", storePath, "File path of the Thing store.")

	lis := listener.CreateServiceListener(ServiceName)

	svc, err := directorykvstore.NewDirectoryKVStoreServer(storePath)
	if err != nil {
		log.Fatalf("Service '%s' failed to start: %s", ServiceName, err)
	}
	logrus.Infof("DirectoryCapnpServer starting on: %s", lis.Addr())
	capnpserver.StartDirectoryCapnpServer(context.Background(), lis, svc)
}
