package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/authz/pkg/aclstore"
	"github.com/wostzone/hub/authz/pkg/authorize"
	"os"
	"path"
)

const Version = `0.3-alpha`

func main() {
	binFolder := path.Dir(os.Args[0])
	homeFolder := path.Dir(binFolder)
	ParseArgs(homeFolder, os.Args[1:])
}

// ParseArgs to handle commandline arguments
func ParseArgs(homeFolder string, args []string) {
	// var err error
	configFolder := path.Join(homeFolder, "config")
	// configFolder := path.Join(homeFolder, "config")
	// genKeys := false
	var optConf struct {
		// commands
		Setrole bool
		Remove  bool
		// arguments
		Clientid string
		Groupid  string
		Role     string
		// options
		Aclfile  string
		Config   string
		Certs    string
		Hostname string
		Output   string
		Pubkey   string
		Iter     int
		Verbose  bool
	}
	usage := `
Usage:
  authz setrole [-v -c configFolder --aclfile=aclfile] <clientID> <groupID> <role>
  authz remove  [-v -c configFolder --aclfile=aclfile] <clientID> [<groupID>]
  authz --help | --version

Commands:
  setrole      Set user role in group
  remove       Remove the user from a group or all groups

Arguments:
  clientID     ID of the client to set the role for
  groupID      group whose resources to allow access to or no group for all
  role         one of viewer, operator, manager, thing, or none to delete

Options:
  --aclfile=AclFile          use a different acl file instead of the default config/` + aclstore.DefaultAclFile + `
  -c --config=ConfigFolder   location of Hub config folder containing the ACL file [default: ` + configFolder + `]
  -h --help                  show this help
  -v --verbose               show info logging
  --version                  show app version
`
	opts, err := docopt.ParseArgs(usage, args, Version)
	if err != nil {
		fmt.Printf("Parse Error: %s\n", err)
		os.Exit(1)
	}

	err = opts.Bind(&optConf)

	if optConf.Verbose {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	if err != nil {
		fmt.Printf("Bind Error: %s\n", err)
		os.Exit(1)
	}
	_ = opts
	if optConf.Setrole {
		fmt.Printf("Set user role in group %s\n", optConf.Groupid)
		err = HandleSetRole(optConf.Config, optConf.Clientid, optConf.Groupid, optConf.Role, optConf.Aclfile)
	} else if optConf.Remove {
		fmt.Printf("Remove user from group\n")
		err = HandleRemove(optConf.Config, optConf.Clientid, optConf.Groupid, optConf.Aclfile)
	} else {
		err = fmt.Errorf("invalid command")
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

// HandleSetRole sets the role of a client in a group.
func HandleSetRole(configFolder string, clientID string, groupID string, role string, aclFile string) error {
	if role != authorize.GroupRoleOperator && role != authorize.GroupRoleViewer &&
		role != authorize.GroupRoleManager && role != authorize.GroupRoleThing && role != authorize.GroupRoleNone {
		err := fmt.Errorf("invalid role '%s'", role)
		return err
	}
	aclFilePath := path.Join(configFolder, aclstore.DefaultAclFile)
	if aclFile != "" {
		// option to specify an acl file wrt home
		aclFilePath = aclFile
	}
	fmt.Printf("Using config folder: %s\n", configFolder)
	fmt.Printf("Using acl file: %s\n", aclFilePath)

	aclStore := aclstore.NewAclFileStore(aclFilePath, "author.main.HandleSetRole")
	err := aclStore.Open()
	if err == nil {
		err = aclStore.SetRole(clientID, groupID, role)
	}
	if err == nil {
		fmt.Printf("Client '%s' role set to '%s' for group '%s'\n", clientID, role, groupID)
	}
	aclStore.Close()
	return err
}

// HandleRemove removes a client from a group
func HandleRemove(configFolder string, clientID string, groupID string, aclFile string) error {
	aclFilePath := path.Join(configFolder, aclstore.DefaultAclFile)
	if aclFile != "" {
		// option to specify an acl file wrt home
		aclFilePath = path.Join(path.Dir(configFolder), aclFile)
	}
	aclStore := aclstore.NewAclFileStore(aclFilePath, "author.main.HandleRemove")
	err := aclStore.Open()
	if err == nil {
		err = aclStore.Remove(clientID, groupID)
	}
	if err == nil {
		fmt.Printf("Client '%s' removed from group '%s'\n", clientID, groupID)
	}
	aclStore.Close()
	return err
}
