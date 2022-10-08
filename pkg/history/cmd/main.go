// Package main with the history store
package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpserver"
	"github.com/hiveot/hub/pkg/history/config"
	"github.com/hiveot/hub/pkg/history/service/mongohs"
)

// DefaultConfigFile is the default configuration file with database settings
const DefaultConfigFile = "history.yaml"

// Start the history store service
func main() {

	homeFolder := filepath.Join(filepath.Dir(os.Args[0]), "../..")
	f := folders.GetFolders(homeFolder, false)

	svcConfig := config.NewHistoryStoreConfig()
	configFile := path.Join(f.Config, DefaultConfigFile)

	// Add commandline option '-c configFile which holds service connection info
	flag.StringVar(&configFile, "c", configFile, "Service configuration with database connection info")
	flag.Parse()

	srvListener := listener.CreateServiceListener(f.Run, history.ServiceName)

	// config file is optional
	configData, err := ioutil.ReadFile(configFile)
	if err == nil {
		err = yaml.Unmarshal(configData, &svcConfig)
		if err != nil {
			logrus.Fatalf("Error reading service configuration file '%s': %v", configFile, err)
		}
	}
	// For now only mongodb is supported
	// This service needs the storage location and name
	svc := mongohs.NewMongoHistoryServer(svcConfig)

	_ = capnpserver.StartHistoryCapnpServer(context.Background(), srvListener, svc)
}
