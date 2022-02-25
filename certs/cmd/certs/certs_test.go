package main

import (
	"os"
	"path"
	"strings"
	"testing"
)

func TestArgs(t *testing.T) {
	wd, _ := os.Getwd()
	homeFolder := path.Join(wd, "../../test")
	os.Chdir(homeFolder)

	// cmdline := "setpassword client1 bob -c ./config"
	cmdline := "certbundle"
	args := strings.Split(cmdline, " ")
	ParseArgs(homeFolder, args)
}

func TestNoArgs(t *testing.T) {

}
