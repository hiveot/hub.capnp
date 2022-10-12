// Package main with the thing directory store
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/logging"
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
	logging.SetLogging("info", "")

	homeFolder := filepath.Join(filepath.Dir(os.Args[0]), "../..")
	f := folders.GetFolders(homeFolder, false)
	storePath := filepath.Join(f.Stores, directory.ServiceName, storeFile)
	flag.StringVar(&storePath, "f", storePath, "File path of the directory store.")
	flag.Parse()

	srvListener := listener.CreateServiceListener(f.Run, directory.ServiceName)
	ctx := context.Background()

	svc, err := directorykvstore.NewDirectoryKVStoreServer(ctx, storePath)
	if err == nil {
		logrus.Infof("DirectoryCapnpServer starting on: %s", srvListener.Addr())
		err = capnpserver.StartDirectoryCapnpServer(ctx, srvListener, svc)
	}
	if err != nil {
		//logrus.Fatalf("Service '%s' failed to start: %s", directory.ServiceName, err)
		msg := fmt.Sprintf("ERROR: Service '%s' failed to start: %s\n", directory.ServiceName, err)
		os.Stderr.Write([]byte(msg))
		//logrus.Fatal(msg)
	}
	logrus.Warningf("Directory ended gracefully")
	os.Stderr.WriteString("---test1. os.stderr dire ended\n")
	os.Stdout.WriteString("---test2. os.stdout dire ended\n")
	logrus.Error("test3 Directory ended gracefully")
	logrus.Fatal("test4 Directory ended gracefully")

	os.Exit(0)
}
