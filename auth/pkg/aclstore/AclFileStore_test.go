package aclstore_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/auth/pkg/aclstore"
	"github.com/wostzone/hub/auth/pkg/authorize"
	"github.com/wostzone/hub/lib/client/pkg/config"
)

// NOTE: this name must match the auth_opt_* filenames in mosquitto.conf.template
// also used in mosquittomgr testing
const aclFileName = "testaclstore.acl" // auth_opt_aclFile
var aclFilePath string
var configFolder string

// TestMain for all auth tests, setup of default folders and filenames
func TestMain(m *testing.M) {
	_ = config.SetLogging("info", "")
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	configFolder = path.Join(homeFolder, "config")

	// Make sure ACL and password files exist
	aclFilePath = path.Join(configFolder, aclFileName)
	fp, _ := os.Create(aclFilePath)
	// fp.WriteString("group1:\n  user1: manager\n")
	_ = fp.Close()

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

func TestSetRoleAndRestart(t *testing.T) {
	user1 := "user1"
	user2 := "user2"
	role1 := authorize.GroupRoleManager
	role2 := authorize.GroupRoleManager
	group1 := "group1"
	group2 := "all"
	aclStore := aclstore.NewAclFileStore(aclFilePath, "TestSetRole")
	err := aclStore.Open()
	assert.NoError(t, err)

	err = aclStore.SetRole(user1, group1, role1)
	err = aclStore.SetRole(user1, group2, role1)
	err = aclStore.SetRole(user2, group2, role2)
	assert.NoError(t, err)

	// stop and reload
	aclStore.Close()
	err = aclStore.Open()
	assert.NoError(t, err)

	// time to reload
	time.Sleep(time.Second)

	groups := aclStore.GetGroups(user1)
	assert.GreaterOrEqual(t, len(groups), 1)
	ur1 := aclStore.GetRole(user1, groups)
	assert.Equal(t, role1, ur1)

	groups = aclStore.GetGroups(user2)
	ur2 := aclStore.GetRole(user2, groups)
	assert.Equal(t, role1, ur2)

	aclStore.Close()
}

func TestRemoveRole(t *testing.T) {
	user1 := "user1"
	role1 := authorize.GroupRoleManager
	group1 := "group1"
	aclStore := aclstore.NewAclFileStore(aclFilePath, "TestSetRole")
	err := aclStore.Open()
	assert.NoError(t, err)

	err = aclStore.SetRole(user1, group1, role1)
	assert.NoError(t, err)

	// clearing role should remove user from the group
	err = aclStore.SetRole(user1, group1, authorize.GroupRoleNone)
	assert.NoError(t, err)

	// needs reload to take effect
	time.Sleep(time.Second)

	groups := aclStore.GetGroups(user1)
	assert.Equal(t, 0, len(groups))

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
	ge := aclstore.IsRoleGreaterEqual(authorize.GroupRoleViewer, authorize.GroupRoleNone)
	assert.True(t, ge)
	ge = aclstore.IsRoleGreaterEqual(authorize.GroupRoleNone, authorize.GroupRoleViewer)
	assert.False(t, ge)

	ge = aclstore.IsRoleGreaterEqual(authorize.GroupRoleEditor, authorize.GroupRoleViewer)
	assert.True(t, ge)
	ge = aclstore.IsRoleGreaterEqual(authorize.GroupRoleViewer, authorize.GroupRoleEditor)
	assert.False(t, ge)

	ge = aclstore.IsRoleGreaterEqual(authorize.GroupRoleManager, authorize.GroupRoleEditor)
	assert.True(t, ge)
	ge = aclstore.IsRoleGreaterEqual(authorize.GroupRoleEditor, authorize.GroupRoleManager)
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
	badAclFile := path.Join(configFolder, "badaclfile.acl")
	fp, _ := os.Create(badAclFile)
	fp.WriteString("This is not a valid acl file\nParsing should fail.")
	as := aclstore.NewAclFileStore(badAclFile, "TestBadAclFile")
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
