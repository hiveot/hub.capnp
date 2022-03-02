package main

import (
	"bufio"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/authn/pkg/authenticate"
	"github.com/wostzone/hub/authn/pkg/unpwstore"
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
		// arguments
		Loginid string
		Passwd  string
		// options
		Config  string
		Iter    int
		Verbose bool
	}
	usage := `
Usage:
  authn setpassword [-v -c configFolder] [-i iterations] <loginID> [<passwd>] 
  authn --help | --version

Commands:
  setpassword  Set user password

Arguments:
  loginID      used as the certificate CN, login name and certificate filename (loginID-cert.pem)
  passwd       optional passwd or leave empty to prompt

Options:
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
		err = HandleSetPasswd(optConf.Config, optConf.Loginid, optConf.Passwd, optConf.Iter)
	} else {
		err = fmt.Errorf("invalid command")
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

// HandleSetPasswd sets the login name and password for a consumer
func HandleSetPasswd(configFolder string, username string, passwd string, iterations int) error {
	var pwHash string
	var err error

	reader := bufio.NewReader(os.Stdin)
	unpwFilePath := path.Join(configFolder, unpwstore.DefaultPasswordFile)
	unpwStore := unpwstore.NewPasswordFileStore(unpwFilePath, "authn.main.HandleSetPasswd")
	err = unpwStore.Open()
	if err == nil {
		if passwd == "" {
			fmt.Printf("\nNew Password: ")
			passwd, err = reader.ReadString('\n')
			passwd = strings.Replace(passwd, "\n", "", -1)
		}
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
