package folders

import (
	"os"
	"path/filepath"
)

type AppFolders struct {
	Bin      string // Application binary folder, eg launcher, cli, ...
	Services string // Services and plugin folder
	Home     string // Home folder, default this is the parent of bin, config, certs and logs
	Config   string // Config folder with application and service yaml configuration files
	Certs    string // Certificates and keys
	Logs     string // Logging output
	Run      string // PID and sockets folder.
	Stores   string // Root of the service stores
}

// GetFolders returns the application folders for use by the Hub.
//
// The default 'user based' structure is:
//   home
//     |- bin                Application binaries, cli and launcher
//         |- services       Service and plugin binaries
//     |- config             Service configuration yaml files
//     |- certs              CA and service certificates
//     |- logs               Logging output
//     |- run                PID files and sockets
//     |- stores
//         |- {service}      Store for service
//
// The system based folder structure is:
//   /opt/hiveot               Application binaries, cli and launcher
//         |-- services        Service and plugin binaries
//   /etc/hiveot/conf.d        Service configuration yaml files
//   /etc/hiveot/certs         CA and service certificates
//   /var/log/hiveot           Logging output
//   /run/hiveot               PID files and sockets
//   /var/lib/hiveot/{service} Storage of service
//
// This uses os.Args[0] application path to determine the bin folder, and home as parentFolder
// The services folder is a subdirectory of bin
//  homeFolder is optional in order to override the paths. Use "" for defaults
func GetFolders(homeFolder string, useSystem bool) AppFolders {
	// note that filepath should support windows
	if homeFolder == "" {
		// by default, home is the parent of bin
		homeFolder = filepath.Join(filepath.Dir(os.Args[0]), "..")
	}
	binFolder := filepath.Join(homeFolder, "bin")
	servicesFolder := filepath.Join(binFolder, "services")
	configFolder := filepath.Join(homeFolder, "config")
	certsFolder := filepath.Join(homeFolder, "certs")
	logsFolder := filepath.Join(homeFolder, "logs")
	runFolder := filepath.Join(homeFolder, "run")
	storesFolder := filepath.Join(homeFolder, "store")

	if useSystem {
		homeFolder = filepath.Join("/opt", "hiveot")
		binFolder = homeFolder
		servicesFolder = filepath.Join(binFolder, "services")
		configFolder = filepath.Join("/etc", "hiveot", "conf.d")
		certsFolder = filepath.Join("/etc", "hiveot", "certs")
		logsFolder = filepath.Join("/var", "log", "hiveot")
		runFolder = filepath.Join("/run", "hiveot")
		storesFolder = filepath.Join("/var", "lib", "hiveot")
	}

	return AppFolders{
		Bin:      binFolder,
		Services: servicesFolder,
		Home:     homeFolder,
		Config:   configFolder,
		Certs:    certsFolder,
		Logs:     logsFolder,
		Run:      runFolder,
		Stores:   storesFolder,
	}
}
