package main

import (
	"os"
	"path"
	"strings"
	"testing"
)

func TestArgs(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "wost-certs-test")
	certsFolder := path.Join(tempFolder, "certs")
	_ = os.MkdirAll(certsFolder, 0700)
	_ = os.Chdir(tempFolder)

	cmdline := "certbundle"
	args := strings.Split(cmdline, " ")
	ParseArgs(tempFolder, args)
}

func TestNoArgs(t *testing.T) {

}
