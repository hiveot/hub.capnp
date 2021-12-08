package authorize_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/auth/pkg/aclstore"
	"github.com/wostzone/hub/auth/pkg/authorize"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/td"
	"github.com/wostzone/hub/lib/serve/pkg/certsetup"
)

const testDevice1 = "device1"
const aclFileName = "test-ah.acl" // auth_opt_aclFile
var aclFilePath string

const unpwFileName = "test.passwd"

var unpwFilePath string

// TestMain for all auth tests, setup of default folders and filenames
func TestMain(m *testing.M) {
	config.SetLogging("info", "")
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	configFolder := path.Join(homeFolder, "config")

	// Make sure ACL and password files exist
	aclFilePath = path.Join(configFolder, aclFileName)
	fp, _ := os.Create(aclFilePath)
	fp.Close()
	unpwFilePath = path.Join(configFolder, unpwFileName)
	fp, _ = os.Create(unpwFilePath)
	fp.Close()
	// creating these files takes a bit of time,
	time.Sleep(time.Second)

	res := m.Run()
	os.Exit(res)
}

// Create auth handler with empty username/pw and acl list
func createEmptyTestAuthHandler() *authorize.Authorizer {
	fp, _ := os.Create(aclFilePath)
	fp.Close()
	fp, _ = os.Create(unpwFilePath)
	fp.Close()
	aclStore := aclstore.NewAclFileStore(aclFilePath, "createEmptyTestAuthHandler")
	ah := authorize.NewAuthorizer(aclStore)
	return ah
}

func TestAuthHandlerStartStop(t *testing.T) {
	logrus.Infof("---TestAuthHandlerStartStop---")
	ah := createEmptyTestAuthHandler()
	err := ah.Start()
	time.Sleep(time.Second * 1)
	assert.NoError(t, err)
	ah.Stop()
}

func TestAuthHandlerStartStopNoPw(t *testing.T) {
	logrus.Infof("---TestAuthHandlerStartStopNoPw---")
	aclStore := aclstore.NewAclFileStore(aclFilePath, "TestAuthHandlerStartStopNoPw")
	ah := authorize.NewAuthorizer(aclStore)
	err := ah.Start()
	time.Sleep(time.Second * 1)
	assert.NoError(t, err)
	ah.Stop()
}
func TestAuthHandlerBadStart(t *testing.T) {
	logrus.Infof("---TestAuthHandlerBadStart---")
	aclStore := aclstore.NewAclFileStore("/bad/aclstore/path", "TestAuthHandlerBadStart")
	ah := authorize.NewAuthorizer(aclStore)

	// opening the acl store should fail
	err := ah.Start()
	assert.Error(t, err)
	ah.Stop()

	// missing store should not panic
	ah = authorize.NewAuthorizer(nil)
	err = ah.Start()
	assert.Error(t, err)
}

func TestIsPublisher(t *testing.T) {
	logrus.Infof("---TestIsPublisher---")

	thingID1 := "urn:zone:" + testDevice1 + ":sensor1:temperature"
	thingID2 := "urn:zone:" + testDevice1 + ":sensor1"
	thingID3 := "urn:zone:" + testDevice1 + ""
	ah := createEmptyTestAuthHandler()
	ah.Start()

	isPublisher := ah.IsPublisher(testDevice1, thingID1)
	assert.True(t, isPublisher)
	isPublisher = ah.IsPublisher(testDevice1, thingID2)
	assert.False(t, isPublisher)
	isPublisher = ah.IsPublisher(testDevice1, thingID3)
	assert.False(t, isPublisher)
	ah.Stop()
}

func TestHasPermission(t *testing.T) {
	logrus.Infof("---TestHasPermission---")

	ah := createEmptyTestAuthHandler()
	ah.Start()
	// read permission
	hasPerm := ah.VerifyRolePermission(authorize.GroupRoleThing, false, td.MessageTypeTD)
	assert.True(t, hasPerm)
	hasPerm = ah.VerifyRolePermission(authorize.GroupRoleEditor, false, td.MessageTypeTD)
	assert.True(t, hasPerm)
	hasPerm = ah.VerifyRolePermission(authorize.GroupRoleViewer, false, td.MessageTypeTD)
	assert.True(t, hasPerm)
	hasPerm = ah.VerifyRolePermission(authorize.GroupRoleManager, false, td.MessageTypeTD)
	assert.True(t, hasPerm)

	hasPerm = ah.VerifyRolePermission(authorize.GroupRoleNone, false, td.MessageTypeTD)
	assert.False(t, hasPerm)
	// write permission
	hasPerm = ah.VerifyRolePermission(authorize.GroupRoleThing, true, td.MessageTypeTD)
	assert.True(t, hasPerm)
	hasPerm = ah.VerifyRolePermission(authorize.GroupRoleViewer, true, td.MessageTypeTD)
	assert.False(t, hasPerm)

	ah.Stop()

}

func TestCheckDeviceAuthorization(t *testing.T) {
	logrus.Infof("---TestCheckDeviceAuthorization---")
	aclStore := aclstore.NewAclFileStore(aclFilePath, "TestCheckDeviceAuthorization")

	ah := authorize.NewAuthorizer(aclStore)
	ah.Start()

	group1 := "group1"
	userName := "pub1"
	thingID1 := "urn:zone1:pub1:device1:sensor1"
	thingID2 := "urn:zone1:pub2:device1:sensor1"
	writing := true
	msgType := td.MessageTypeTD

	// publishers can publish to things with thingID that contains the publisher
	authorized := ah.VerifyAuthorization(userName, certsetup.OUIoTDevice, thingID1, writing, msgType)
	assert.True(t, authorized)
	// publishers can not publish to things from another publisher
	authorized = ah.VerifyAuthorization(userName, certsetup.OUIoTDevice, thingID2, writing, msgType)
	assert.False(t, authorized)

	// plugins can do whatever
	authorized = ah.VerifyAuthorization("", certsetup.OUPlugin, thingID1, writing, msgType)
	assert.True(t, authorized)

	// users without role cannot publish
	authorized = ah.VerifyAuthorization(userName, "", thingID1, writing, msgType)
	assert.False(t, authorized)

	// users cannot publish ... unless their role allows it
	authorized = ah.VerifyAuthorization("user1", "", thingID1, writing, msgType)
	assert.False(t, authorized)
	grps := aclStore.GetGroups(thingID1)
	assert.Zero(t, len(grps))
	// viewer roles cannot publish
	aclStore.SetRole(thingID1, group1, authorize.GroupRoleThing)
	aclStore.SetRole("user1", group1, authorize.GroupRoleViewer)
	time.Sleep(time.Millisecond * 200) // reload

	authorized = ah.VerifyAuthorization("user1", "", thingID1, writing, msgType)
	assert.False(t, authorized)
	// editor role can control thing with actions
	aclStore.SetRole("user1", group1, authorize.GroupRoleEditor)
	time.Sleep(time.Millisecond * 200) // reload
	authorized = ah.VerifyAuthorization("user1", "", thingID1, writing, td.MessageTypeAction)
	assert.True(t, authorized)
	ah.Stop()

}