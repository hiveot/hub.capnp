package auth_test

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/pkg/auth"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
)

const testDevice1 = "device1"

// Create auth handler with empty username/pw and acl list
func createEmptyTestAuthHandler() *auth.AuthHandler {
	fp, _ := os.Create(unpwFilePath)
	fp.Close()
	unpwStore := auth.NewPasswordFileStore(unpwFilePath)
	aclStore := auth.NewAclFileStore(aclFilePath)
	ah := auth.NewAuthHandler(aclStore, unpwStore)
	return ah
}

func TestAuthHandlerStartStop(t *testing.T) {
	logrus.Infof("---TestStartStop---")
	ah := createEmptyTestAuthHandler()
	err := ah.Start()
	time.Sleep(time.Second * 1)
	assert.NoError(t, err)
	ah.Stop()
}

func TestAuthHandlerBadStart(t *testing.T) {
	logrus.Infof("---TestBadStart---")
	unpwStore := auth.NewPasswordFileStore(unpwFilePath)
	aclStore := auth.NewAclFileStore("/bad/aclstore/path")
	ah := auth.NewAuthHandler(aclStore, unpwStore)

	// opening the acl store should fail
	err := ah.Start()
	assert.Error(t, err)
	ah.Stop()
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
	hasPerm := ah.HasPermission(auth.GroupRoleThing, false, hubclient.MessageTypeTD)
	assert.NoError(t, hasPerm)
	hasPerm = ah.HasPermission(auth.GroupRoleEditor, false, hubclient.MessageTypeTD)
	assert.NoError(t, hasPerm)
	hasPerm = ah.HasPermission(auth.GroupRoleViewer, false, hubclient.MessageTypeTD)
	assert.NoError(t, hasPerm)
	hasPerm = ah.HasPermission(auth.GroupRoleManager, false, hubclient.MessageTypeTD)
	assert.NoError(t, hasPerm)

	hasPerm = ah.HasPermission(auth.GroupRoleNone, false, hubclient.MessageTypeTD)
	assert.Error(t, hasPerm)
	// write permission
	hasPerm = ah.HasPermission(auth.GroupRoleThing, true, hubclient.MessageTypeTD)
	assert.NoError(t, hasPerm)
	hasPerm = ah.HasPermission(auth.GroupRoleViewer, true, hubclient.MessageTypeTD)
	assert.Error(t, hasPerm)

	ah.Stop()

}

func TestCheckDeviceAuthorization(t *testing.T) {
	logrus.Infof("---TestCheckDeviceAuthorization---")
	unpwStore := auth.NewPasswordFileStore(unpwFilePath)
	aclStore := auth.NewAclFileStore(aclFilePath)

	ah := auth.NewAuthHandler(aclStore, unpwStore)
	ah.Start()

	group1 := "group1"
	userName := "pub1"
	thingID1 := "urn:zone1:pub1:device1:sensor1"
	thingID2 := "urn:zone1:pub2:device1:sensor1"
	writing := true
	msgType := hubclient.MessageTypeTD

	// publishers can publish to things with thingID that contains the publisher
	err := ah.CheckAuthorization(userName, certsetup.OUIoTDevice, thingID1, writing, msgType)
	assert.NoError(t, err)
	// publishers can not publish to things from another publisher
	err = ah.CheckAuthorization(userName, certsetup.OUIoTDevice, thingID2, writing, msgType)
	assert.Error(t, err)

	// plugins can do whatever
	err = ah.CheckAuthorization("", certsetup.OUPlugin, thingID1, writing, msgType)
	assert.NoError(t, err)

	// users without role cannot publish
	err = ah.CheckAuthorization(userName, "", thingID1, writing, msgType)
	assert.Error(t, err)

	// users cannot publish ... unless their role allows it
	err = ah.CheckAuthorization("user1", "", thingID1, writing, msgType)
	assert.Error(t, err)
	grps := aclStore.GetGroups(thingID1)
	assert.Zero(t, len(grps))
	// viewer roles cannot publish
	aclStore.SetRole(thingID1, group1, auth.GroupRoleThing)
	aclStore.SetRole("user1", group1, auth.GroupRoleViewer)
	time.Sleep(time.Millisecond * 200) // reload

	err = ah.CheckAuthorization("user1", "", thingID1, writing, msgType)
	assert.Error(t, err)
	// editor role can
	aclStore.SetRole("user1", group1, auth.GroupRoleEditor)
	time.Sleep(time.Millisecond * 200) // reload
	err = ah.CheckAuthorization("user1", "", thingID1, writing, msgType)
	assert.NoError(t, err)
	ah.Stop()

}

func TestUnpwMatch(t *testing.T) {
	logrus.Infof("---TestUnpwMatch---")
	userName := "user1" // as in test file
	password := "user1"

	ah := createEmptyTestAuthHandler()
	ah.Start()

	// add the user to the password file
	unpwStore2 := auth.NewPasswordFileStore(unpwFilePath)
	unpwStore2.Open()
	pwHash, err := auth.CreatePasswordHash(password, auth.PWHASH_ARGON2id, 0)
	assert.NoError(t, err)
	unpwStore2.SetPasswordHash(userName, pwHash)
	unpwStore2.Close()
	time.Sleep(time.Millisecond * 200)

	match := ah.CheckUsernamePassword(userName, password)
	assert.True(t, match)

	ah.Stop()
}

func TestUnpwNoMatch(t *testing.T) {
	logrus.Infof("---TestUnpwNoMatch---")
	ah := createEmptyTestAuthHandler()
	ah.Start()

	user1 := "user1" // user 1 exists in test file
	password := "user1"

	match := ah.CheckUsernamePassword("notauser", password)
	assert.False(t, match)

	match = ah.CheckUsernamePassword(user1, "badpassword")
	assert.False(t, match)

	ah.Stop()

}

func TestBCrypt(t *testing.T) {
	logrus.Infof("---TestBCrypt---")
	var password1 = "password1"
	ah := createEmptyTestAuthHandler()
	err := ah.Start()
	assert.NoError(t, err)
	hash, err := auth.CreatePasswordHash(password1, auth.PWHASH_BCRYPT, 0)
	assert.NoError(t, err)
	match := ah.VerifyPasswordHash(hash, password1, auth.PWHASH_BCRYPT)
	assert.True(t, match)
	ah.Stop()
}
