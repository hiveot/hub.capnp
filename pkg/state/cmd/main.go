package main

import (
	"context"
	"flag"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/state"
	"github.com/hiveot/hub/pkg/state/capnpserver"
	"github.com/hiveot/hub/pkg/state/service/statekvstore"
)

var binFolder string
var homeFolder string

// Start the launcher service
func main() {
	var err error
	var svc state.IState
	var ctx = context.Background()

	logrus.SetLevel(logrus.InfoLevel)
	// this is a service so go 2 levels up
	// FIXME: import the folder structure instead of hard coding it
	homeFolder := filepath.Join(filepath.Dir(os.Args[0]), "../..")
	f := folders.GetFolders(homeFolder, false)
	stateStorePath := path.Join(f.Config, certs.ServiceName+".json")

	// option to override the location of the store. Intended for testing
	flag.StringVar(&stateStorePath, "storePath", stateStorePath, "State store file")
	flag.Parse()

	srvListener := listener.CreateServiceListener(f.Run, certs.ServiceName)

	if err == nil {
		svc, err = statekvstore.NewStateKVStore(stateStorePath)
	}
	if err == nil {
		err = capnpserver.StartStateCapnpServer(ctx, srvListener, svc)
	}
}
