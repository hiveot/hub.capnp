package authhandler_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/core/authhandler"
	"github.com/wostzone/hub/pkg/auth"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

var homeFolder string
var aclFile string

// var pwFile string
var unpwStore *auth.PasswordFileStore
var aclStore *auth.AclFileStore

const testDevice1 = "device1"

// TestMain initializes the test stores
func TestMain(m *testing.M) {
	hubconfig.SetLogging("info", "")
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	configFolder := path.Join(homeFolder, "config")

	aclFile = path.Join(configFolder, "acl-test.yaml")
	os.Create(aclFile) // start with empty file
	aclStore = auth.NewAclFileStore(aclFile)

	unpwFileName := path.Join(configFolder, "unpw-test.conf")
	unpwStore = auth.NewPasswordFileStore(unpwFileName)

	result := m.Run()

	os.Exit(result)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("---TestStartStop---")
	ah := authhandler.NewAuthHandler(aclStore, unpwStore)
	err := ah.Start()
	time.Sleep(time.Second * 1)
	assert.NoError(t, err)
	ah.Stop()
}

func TestBadStart(t *testing.T) {
	logrus.Infof("---TestBadStart---")
	aclStore := auth.NewAclFileStore("/badpath")
	ah := authhandler.NewAuthHandler(aclStore, unpwStore)
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
	ah := authhandler.NewAuthHandler(aclStore, unpwStore)
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

	ah := authhandler.NewAuthHandler(aclStore, unpwStore)
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
	ah := authhandler.NewAuthHandler(aclStore, unpwStore)
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
	ah := authhandler.NewAuthHandler(aclStore, unpwStore)
	ah.Start()

	userName := "user1" // as in test file
	password := "user1"

	err := ah.CheckUsernamePassword(userName, password)
	assert.NoError(t, err)

	ah.Stop()
}

func TestUnpwNoMatch(t *testing.T) {
	logrus.Infof("---TestUnpwNoMatch---")
	ah := authhandler.NewAuthHandler(aclStore, unpwStore)
	ah.Start()

	user1 := "user1" // user 1 exists in test file
	password := "user1"

	err := ah.CheckUsernamePassword("notauser", password)
	assert.Error(t, err)

	err = ah.CheckUsernamePassword(user1, "badpassword")
	assert.Error(t, err)

	ah.Stop()

}

// func TestAuthUnpwdCheck(t *testing.T) {
// 	logrus.Infof("---TestAuthUnpwdCheck---")
// 	username := "user1"
// 	password := "password1"
// 	clientID := "clientID1"
// 	clientIP := "ip"
// 	main.AuthUnpwdCheck(clientID, username, password, clientIP)
// }

// func TestAuthAclCheck(t *testing.T) {
// 	logrus.Infof("---TestAuthAclCheck---")
// 	clientID := "clientID1"
// 	username := "user1"
// 	topic := "things/thingid1/td"
// 	access := main.MOSQ_ACL_SUBSCRIBE
// 	main.AuthAclCheck(clientID, username, topic, access, true)
// }

// func TestAuthPluginCleanup(t *testing.T) {
// 	logrus.Infof("---TestAuthPluginCleanup---")
// 	main.AuthPluginCleanup()
// }
