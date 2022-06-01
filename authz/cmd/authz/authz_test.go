package main

import (
	"os"
	"path"
	"strings"
	"testing"
)

func TestArgs(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "wost-authz-test")
	_ = os.MkdirAll(tempFolder, 0700)
	_ = os.Chdir(tempFolder)
	fp, _ := os.Create("test.acl")
	_ = fp.Close()

	cmdline := "setrole client1 group1 viewer --aclfile=./test.acl"
	args := strings.Split(cmdline, " ")
	ParseArgs(tempFolder, args)
}

func TestNoArgs(t *testing.T) {

}
