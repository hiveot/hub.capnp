package auth_test

import (
	"os"
	"path"
	"testing"

	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

const aclFileName = "test.acl"
const unpwFileName = "test.passwd"

var aclFilePath string
var unpwFilePath string

var configFolder string

// var aclStore *auth.AclFileStore

// TestMain for all auth tests, setup of default folders and filenames
func TestMain(m *testing.M) {
	hubconfig.SetLogging("info", "")
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	configFolder = path.Join(homeFolder, "config")

	// Make sure an ACL file exist
	aclFilePath = path.Join(configFolder, aclFileName)
	unpwFilePath = path.Join(configFolder, unpwFileName)
	fp, _ := os.Create(aclFilePath)
	fp.Close()

	res := m.Run()
	os.Exit(res)
}
