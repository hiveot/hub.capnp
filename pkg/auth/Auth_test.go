package auth_test

import (
	"os"
	"path"
	"testing"

	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// NOTE: these names must match the auth_opt_* filenames in mosquitto.conf.template
// also used in mosquittomgr testing
const aclFileName = "test.acl" // auth_opt_aclFile
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

	// Make sure ACL and password files exist
	aclFilePath = path.Join(configFolder, aclFileName)
	unpwFilePath = path.Join(configFolder, unpwFileName)
	fp, _ := os.Create(aclFilePath)
	fp.Close()
	fp, _ = os.Create(unpwFilePath)
	fp.Close()

	res := m.Run()
	os.Exit(res)
}
