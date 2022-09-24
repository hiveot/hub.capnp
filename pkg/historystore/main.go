// Package main with the history store
package main

import (
	"flag"
	"io/ioutil"
	"path"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/historystore/adapter"
	"github.com/hiveot/hub/pkg/historystore/config"
	"github.com/hiveot/hub/pkg/historystore/mongohs"
)

// DefaultConfigFile is the default configuration file with database settings
const DefaultConfigFile = "historystore.yaml"

// Start the history store service
func main() {
	svcConfig := config.NewHistoryStoreConfig()
	configFile := path.Join(folders.GetFolders("").Config, DefaultConfigFile)

	// Add commandline option '-c configFile which holds service connection info
	flag.StringVar(&configFile, "c", configFile, "Service configuration with database connection info")
	lis := listener.CreateServiceListener(config.ServiceName)

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
	store := mongohs.NewMongoHistoryStoreServer(svcConfig)

	adapter.StartHistoryStoreCapnpAdapter(lis, store)
}
