package main

import (
	"bufio"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/auth/pkg/aclstore"
	"github.com/wostzone/hub/auth/pkg/authenticate"
	"github.com/wostzone/hub/auth/pkg/authorize"
	"github.com/wostzone/hub/auth/pkg/unpwstore"
	"os"
	"path"
	"strings"
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
		Setpassword bool
		Setrole     bool
		// arguments
		Loginid string
		Groupid string
		Role    string
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
  auth setpassword [-v -c configFolder] [-i iterations] <loginID> 
  auth setrole [-v -c configFolder --aclfile=aclfile] <loginID> <groupID> <role>
  auth --help | --version

Commands:
  setpassword  Set user password
  setrole      Set user role in group

Arguments:
  loginID      used as the certificate CN, login name and certificate filename (loginID-cert.pem)
  groupID      group for access control 
  role         one of viewer, editor, manager, thing, or none to delete

Options:
  --aclfile=AclFile          use a different acl file instead of the default config/` + aclstore.DefaultAclFile + `
  -c --config=ConfigFolder   location of Hub config folder [default: ` + configFolder + `]
  -i --iter=iterations       Number of iterations for generating password [default: 10]
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
	if optConf.Setpassword {
		fmt.Printf("Set user password\n")
		err = HandleSetPasswd(optConf.Config, optConf.Loginid, optConf.Iter)
	} else if optConf.Setrole {
		fmt.Printf("Set user role in group\n")
		err = HandleSetRole(optConf.Config, optConf.Loginid, optConf.Groupid, optConf.Role, optConf.Aclfile)
	} else {
		err = fmt.Errorf("invalid command")
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

// HandleSetPasswd sets the login name and password for a consumer
func HandleSetPasswd(configFolder string, username string, iterations int) error {
	var pwHash string
	var err error
	var passwd string
	reader := bufio.NewReader(os.Stdin)
	unpwFilePath := path.Join(configFolder, unpwstore.DefaultPasswordFile)
	unpwStore := unpwstore.NewPasswordFileStore(unpwFilePath, "auth.main.HandleSetPasswd")
	err = unpwStore.Open()
	if err == nil {
		fmt.Printf("\nNew Password: ")
		passwd, err = reader.ReadString('\n')
		passwd = strings.Replace(passwd, "\n", "", -1)
		if err != nil {
			return err
		}
		if passwd == "" {
			return fmt.Errorf("missing password")
		}
		// pwHash, err = authen.CreatePasswordHash(passwd, authen.PWHASH_ARGON2id, uint(iterations))
		pwHash, err = authenticate.CreatePasswordHash(passwd, authenticate.PWHASH_ARGON2id, uint(iterations))
	}
	if err == nil {
		err = unpwStore.SetPasswordHash(username, pwHash)
	}
	if err == nil {
		unpwStore.Close()
		fmt.Printf("Password updated for user %s\n", username)
	}
	return err
}

// HandleSetRole sets the role of a client in a group.
func HandleSetRole(configFolder string, clientID string, groupID string, role string, aclFile string) error {
	if role != authorize.GroupRoleEditor && role != authorize.GroupRoleViewer &&
		role != authorize.GroupRoleManager && role != authorize.GroupRoleThing && role != authorize.GroupRoleNone {
		err := fmt.Errorf("invalid role '%s'", role)
		return err
	}
	aclFilePath := path.Join(configFolder, aclstore.DefaultAclFile)
	if aclFile != "" {
		// option to specify an acl file wrt home
		aclFilePath = path.Join(path.Dir(configFolder), aclFile)
	}
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
