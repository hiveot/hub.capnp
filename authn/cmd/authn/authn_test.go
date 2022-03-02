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
	fp, _ := os.Create("config/hub.passwd")
	fp.Close()

	cmdline := "setpassword -c ./config user1 user1"
	//cmdline := "--version"
	args := strings.Split(cmdline, " ")
	ParseArgs(homeFolder, args)
}

func TestNoArgs(t *testing.T) {

}
