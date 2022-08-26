package folders

import (
	"os"
	"path"
)

type AppFolders struct {
	Bin    string
	Home   string
	Config string
	Certs  string
	Logs   string
}

// GetFolders returns the application folders for use by services
// This uses os.Args[0] application path to determine the bin folder, and home as parentFolder
// homeFolder is optional in order to override the paths. Use "" for defaults
func GetFolders(homeFolder string) AppFolders {
	binFolder := path.Dir(os.Args[0])
	if homeFolder == "" {
		homeFolder = path.Dir(binFolder)
	}
	configFolder := path.Join(homeFolder, "config")
	certFolder := path.Join(homeFolder, "certs")
	logsFolder := path.Join(homeFolder, "logs")

	return AppFolders{
		Bin:    binFolder,
		Home:   homeFolder,
		Config: configFolder,
		Certs:  certFolder,
		Logs:   logsFolder,
	}
}
