package aclstore_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/pkg/aclstore"
	"github.com/wostzone/hub/pkg/auth"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// NOTE: this name must match the auth_opt_* filenames in mosquitto.conf.template
// also used in mosquittomgr testing
const aclFileName = "testaclstore.acl" // auth_opt_aclFile
var aclFilePath string
var configFolder string

// TestMain for all auth tests, setup of default folders and filenames
func TestMain(m *testing.M) {
	hubconfig.SetLogging("info", "")
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	configFolder = path.Join(homeFolder, "config")

	// Make sure ACL and password files exist
	aclFilePath = path.Join(configFolder, aclFileName)
	fp, _ := os.Create(aclFilePath)
	fp.Close()

	res := m.Run()
	os.Exit(res)
}

func TestOpenCloseAclStore(t *testing.T) {
	aclStore := aclstore.NewAclFileStore(aclFilePath, "TestOpenCloseAclStore")
	err := aclStore.Open()
	assert.NoError(t, err)
	time.Sleep(time.Second * 1)
	assert.NoError(t, err)
	aclStore.Close()
}

func TestSetRole(t *testing.T) {
	// as := auth.NewAclStoreFile(aclFile)
	user1 := "user1"
	role1 := auth.GroupRoleManager
	group1 := "group1"
	aclStore := aclstore.NewAclFileStore(aclFilePath, "TestSetRole")
	err := aclStore.Open()
	assert.NoError(t, err)

	err = aclStore.SetRole(user1, group1, role1)
	assert.NoError(t, err)

	// time to reload
	time.Sleep(time.Second)

	groups := aclStore.GetGroups(user1)
	assert.GreaterOrEqual(t, len(groups), 1)

	role := aclStore.GetRole(user1, groups)
	assert.Equal(t, auth.GroupRoleManager, role)

	aclStore.Close()
}

func TestWriteAclToTempFail(t *testing.T) {
	aclStore := aclstore.NewAclFileStore(aclFilePath, "TestWriteAclToTempFail")
	acls := make(map[string]aclstore.AclGroup)

	err := aclStore.Open()
	assert.NoError(t, err)
	_, err = aclstore.WriteAclsToTempFile("/badfolder", acls)
	assert.Error(t, err)
	aclStore.Close()
}

func TestCompareRoles(t *testing.T) {
	ge := aclstore.IsRoleGreaterEqual(auth.GroupRoleViewer, auth.GroupRoleNone)
	assert.True(t, ge)
	ge = aclstore.IsRoleGreaterEqual(auth.GroupRoleNone, auth.GroupRoleViewer)
	assert.False(t, ge)

	ge = aclstore.IsRoleGreaterEqual(auth.GroupRoleEditor, auth.GroupRoleViewer)
	assert.True(t, ge)
	ge = aclstore.IsRoleGreaterEqual(auth.GroupRoleViewer, auth.GroupRoleEditor)
	assert.False(t, ge)

	ge = aclstore.IsRoleGreaterEqual(auth.GroupRoleManager, auth.GroupRoleEditor)
	assert.True(t, ge)
	ge = aclstore.IsRoleGreaterEqual(auth.GroupRoleEditor, auth.GroupRoleManager)
	assert.False(t, ge)

}

func TestMissingAclFile(t *testing.T) {
	as := aclstore.NewAclFileStore("missingaclfile", "TestMissingAclFile")
	err := as.Open()
	assert.Error(t, err)
	as.Close()

}

func TestBadAclFile(t *testing.T) {
	// loading the hub-bad.yaml should fail as it isn't a valid yaml file
	as := aclstore.NewAclFileStore(path.Join(configFolder, "mosquitto.conf.template"), "TestBadAclFile")
	err := as.Open()
	assert.Error(t, err)
	as.Close()
}

func TestFailWriteFile(t *testing.T) {
	as := aclstore.NewAclFileStore("/root/nopermissions", "TestFailWriteFile")

	err := as.Open()
	assert.Error(t, err)

	// err = os.Chmod(aclFile, 0400)
	// assert.NoError(t, err)

	// err = aclStore.SetRole("user1", "group1", "somerole")
	// assert.Error(t, err)
	// os.Remove(aclFile)
	as.Close()
}
