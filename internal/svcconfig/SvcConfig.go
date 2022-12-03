package svcconfig

import (
	"flag"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// LoadServiceConfig Load a configuration file from the config folder and applies commandline options.
// This invokes Fatal if the configuration file is invalid, or required but not found.
//
// This invokes flag.Parse(). Flag commandline options added are:
//
//	 -c configFile
//	 --home directory
//	 --certs directory
//	 --services directory
//	 --logs directory
//	 --run directory
//
//	If a 'cfg' interface is provided, the configuration is loaded from file and parsed as yaml.
//
//	serviceName is used for the configuration file with the '.yaml' extension
//	required returns an error if the configuration file doesn't exist
//	cfg is the interface to the configuration object. nil to ignore configuration and just load the folders.
func LoadServiceConfig(serviceName string, required bool, cfg interface{}) AppFolders {
	// run the commandline options
	var err error
	var cfgData []byte
	var homeFolder = ""
	var certsFolder = ""
	var runFolder = ""
	var servicesFolder = ""
	var logsFolder = ""
	var storesFolder = ""

	f := GetFolders(homeFolder, false)
	cfgFile := path.Join(f.Config, serviceName+".yaml")
	if cfg != nil {
		flag.StringVar(&cfgFile, "c", cfgFile, "Service config file")
	}
	flag.StringVar(&homeFolder, "home", f.Home, "Application home directory")
	flag.StringVar(&certsFolder, "certs", f.Certs, "Certificates directory")
	flag.StringVar(&servicesFolder, "services", f.Services, "Application services directory")
	flag.StringVar(&logsFolder, "logs", f.Logs, "Service log files directory")
	flag.StringVar(&runFolder, "run", f.Run, "Runtime directory for sockets and pid files")
	flag.StringVar(&storesFolder, "stores", f.Stores, "Storage directory")
	flag.Parse()

	// homefolder is special as it overrides all other folders
	// detect the override by comparing original folder with assigned folder
	f2 := GetFolders(homeFolder, false)
	if certsFolder != f.Certs {
		f2.Certs = certsFolder
	}
	if servicesFolder != f.Services {
		f2.Services = servicesFolder
	}
	if logsFolder != f.Logs {
		f2.Logs = logsFolder
	}
	if runFolder != f.Run {
		f2.Run = runFolder
	}
	if storesFolder != f.Stores {
		f2.Stores = storesFolder
	}

	// ignore configuration file if no destination interface is given
	if cfg != nil {
		cfgData, err = os.ReadFile(cfgFile)
		if err == nil {
			logrus.Infof("Loaded configuration file: %s", cfgFile)
			err = yaml.Unmarshal(cfgData, cfg)
			if err != nil {
				logrus.Fatalf("Loading configuration file '%s' failed with: %s", cfgFile, err)
			}
		} else if !required {
			logrus.Infof("Configuration file '%s' not found. Ignored.", cfgFile)
			err = nil
		} else {
			logrus.Fatalf("Configuration file '%s' not found but required.", cfgFile)
		}
	}
	return f2
}
