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

	// cmdline := "certbundle"
	// cmdline := "clientcert client1"
	// cmdline := "setpassword client1 bob -c ./config"
	cmdline := "setrole client1 group1 viewer"
	args := strings.Split(cmdline, " ")
	ParseArgs(homeFolder, args)
}