package svcconfig

import (
	"flag"
	"io/ioutil"
	"path"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/hiveot/hub/internal/folders"
)

// LoadServiceConfig Load a configuration file from the config folder and apply commandline options.
// Flag commandline options added are:
//   -c configFile
//   --home directory
//   --certs directory
//   --services directory
//   --logs directory
//   --run directory
//
//  After the configuration is loaded from file, the commandline options are parsed.
//  Services can set a flag on the provided cfg to override the defaults or the loaded config.
//
//  f default folders. Can be overridden with commandline options.
//  serviceName is used for the configuration file with the '.yaml' extension
//  required returns an error if the configuration file doesn't exist
//  cfg is the interface to the configuration object, overridden by commandline options
// This returns an error if the yaml file has a typo in it
func LoadServiceConfig(f folders.AppFolders, serviceName string, required bool, cfg interface{}) (folders.AppFolders, error) {
	// run the commandline options
	cfgFile := path.Join(f.Config, serviceName+".yaml")
	flag.StringVar(&cfgFile, "c", cfgFile, "Service config file")
	flag.StringVar(&f.Home, "home", f.Home, "Application home directory")
	flag.StringVar(&f.Certs, "certs", f.Certs, "Certificates directory")
	flag.StringVar(&f.Services, "services", f.Services, "Application services directory")
	flag.StringVar(&f.Logs, "logs", f.Logs, "Service log files directory")
	flag.StringVar(&f.Run, "run", f.Run, "Runtime directory for sockets and pid files")
	flag.Parse()

	cfgData, err := ioutil.ReadFile(cfgFile)
	if err == nil {
		logrus.Infof("Loaded configuration file: %s", cfgFile)
		err = yaml.Unmarshal(cfgData, cfg)
	} else if !required {
		logrus.Infof("Configuration file '%s' not found. Ignored.", cfgFile)
		err = nil
	} else {
		logrus.Errorf("Configuration file '%s' not found but required.", cfgFile)
	}
	return f, err
}
