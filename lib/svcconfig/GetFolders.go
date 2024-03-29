package svcconfig

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type AppFolders struct {
	Bin        string // Application binary folder, eg launcher, cli, ...
	Bindings   string // Bindings folder, use "" to ignore and have everything in services
	Services   string // Services folder
	Home       string // Home folder, default this is the parent of bin, config, certs and logs
	Config     string // Config folder with application and service yaml configuration files
	ConfigFile string
	Certs      string // Certificates and keys
	Logs       string // Logging output
	Run        string // PID and sockets folder.
	Stores     string // Root of the service stores
	SocketPath string // default location of service UDS listening socket, if used
}

// GetFolders returns the application folders for use by the Hub.
//
// The default 'user based' structure is:
//
//	home
//	  |- bin                Application cli and launcher binaries
//	      |- bindings       Protocol binding binaries
//	      |- services       Service binaries
//	  |- config             Service and binding configuration yaml files
//	  |- certs              CA and service certificates
//	  |- logs               Logging output
//	  |- run                PID files and sockets
//	  |- stores
//	      |- {service}      Store for service
//
// The system based folder structure is:
//
//	/opt/hiveot/bin            Application binaries, cli and launcher
//	             |-- bindings  Protocol binding binaries
//	             |-- services  Service binaries
//	/etc/hiveot/conf.d         Service configuration yaml files
//	/etc/hiveot/certs          CA and service certificates
//	/var/log/hiveot            Logging output
//	/run/hiveot                PID files and sockets
//	/var/lib/hiveot/{service}  Storage of service
//
// This uses os.Args[0] application path to determine the services folder. The home folder is two
// levels up from the services folder.
//
//	homeFolder is optional in order to override the auto detected paths. Use "" for defaults.
func GetFolders(homeFolder string, useSystem bool) AppFolders {
	// note that filepath should support windows
	if homeFolder == "" {
		cwd := filepath.Dir(os.Args[0])
		if strings.HasSuffix(cwd, "services") {
			homeFolder = filepath.Join(cwd, "..", "..")
		} else if strings.HasSuffix(cwd, "bindings") {
			homeFolder = filepath.Join(cwd, "..", "..")
		} else if strings.HasSuffix(cwd, "bin") {
			homeFolder = filepath.Join(cwd, "..")
		} else {
			// not sure where home is. For now use the parent
			homeFolder = filepath.Join(cwd, "..")
		}
	}
	//logrus.Infof("homeFolder is '%s", homeFolder)
	binFolder := filepath.Join(homeFolder, "bin")
	bindingsFolder := filepath.Join(binFolder, "bindings")
	servicesFolder := filepath.Join(binFolder, "services")
	configFolder := filepath.Join(homeFolder, "config")
	certsFolder := filepath.Join(homeFolder, "certs")
	logsFolder := filepath.Join(homeFolder, "logs")
	runFolder := filepath.Join(homeFolder, "run")
	storesFolder := filepath.Join(homeFolder, "stores")

	if useSystem {
		homeFolder = filepath.Join("/opt", "hiveot")
		binFolder = homeFolder
		bindingsFolder = filepath.Join(binFolder, "bindings")
		servicesFolder = filepath.Join(binFolder, "services")
		configFolder = filepath.Join("/etc", "hiveot", "conf.d")
		certsFolder = filepath.Join("/etc", "hiveot", "certs")
		logsFolder = filepath.Join("/var", "log", "hiveot")
		runFolder = filepath.Join("/run", "hiveot")
		storesFolder = filepath.Join("/var", "lib", "hiveot")
	}

	return AppFolders{
		Bin:      binFolder,
		Bindings: bindingsFolder,
		Services: servicesFolder,
		Home:     homeFolder,
		Config:   configFolder,
		Certs:    certsFolder,
		Logs:     logsFolder,
		Run:      runFolder,
		Stores:   storesFolder,
	}
}

// LoadConfig loads a configuration file from the config folder
func (f *AppFolders) LoadConfig(cfg interface{}) error {
	cfgData, err := os.ReadFile(f.ConfigFile)
	if err == nil {
		logrus.Infof("Loaded configuration file: %s", f.ConfigFile)
		err = yaml.Unmarshal(cfgData, cfg)
		if err != nil {
			logrus.Fatalf("Loading configuration file '%s' failed with: %s", f.ConfigFile, err)
		}
	} else {
		logrus.Infof("Configuration file '%s' not found. Ignored.", f.ConfigFile)
		err = nil
	}

	return err
}
