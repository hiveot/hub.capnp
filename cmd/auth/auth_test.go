package main

import (
	"os"
	"path"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	wd, _ := os.Getwd()
	homeFolder := path.Join(wd, "../../test")
	os.Chdir(homeFolder)
	fp, _ := os.Create("config/test.acl")
	fp.Close()

	// cmdline := "certbundle"
	// cmdline := "clientcert client1"
	// cmdline := "setpassword client1 bob -c ./config"
	cmdline := "setrole client1 group1 viewer --aclfile=config/test.acl"
	args := strings.Split(cmdline, " ")
	ParseArgs(homeFolder, args)
}
