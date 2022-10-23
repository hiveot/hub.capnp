package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wostzone/hub/authn/pkg/unpwstore"
)

func TestArgs(t *testing.T) {
	tempFolder := path.Join(os.TempDir(), "wost-authn-test")
	os.MkdirAll(tempFolder, 0700)

	passwdFile := path.Join(tempFolder, unpwstore.DefaultPasswordFile)
	fp, _ := os.Create(passwdFile)
	fp.Close()

	cmdline := fmt.Sprintf("setpassword -c %s user1 user1", tempFolder)
	//cmdline := "--version"
	args := strings.Split(cmdline, " ")
	err := ParseArgs(homeFolder, args)
	assert.NoError(t, err)
}

func TestNoArgs(t *testing.T) {

}
