package auth_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/core/auth"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

var homeFolder string
var aclFile string

const testDevice1 = "device1"
const testUser1 = "user1"

// TestMain uses the project test folder as the home folder and generates test certificates
func TestMain(m *testing.M) {
	hubconfig.SetLogging("info", "")
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	configFolder := path.Join(homeFolder, "config")
	aclFile = path.Join(configFolder, "acl-test.yaml")

	result := m.Run()

	os.Exit(result)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("---TestStartStop---")
	ah := auth.NewAuthHandler(auth.NewAclStoreFile(aclFile))
	err := ah.Start()
	time.Sleep(time.Second * 1)
	assert.NoError(t, err)
	ah.Stop()
}

func TestIsPublisher(t *testing.T) {
	logrus.Infof("---TestIsPublisher---")
	thingID1 := "urn:zone:" + testDevice1 + ":sensor1:temperature"
	thingID2 := "urn:zone:" + testDevice1 + ":sensor1"
	thingID3 := "urn:zone:" + testDevice1 + ""
	ah := auth.NewAuthHandler(auth.NewAclStoreFile(aclFile))
	ah.Start()

	isPublisher := ah.IsPublisher(testDevice1, thingID1)
	assert.True(t, isPublisher)
	isPublisher = ah.IsPublisher(testDevice1, thingID2)
	assert.False(t, isPublisher)
	isPublisher = ah.IsPublisher(testDevice1, thingID3)
	assert.False(t, isPublisher)
	ah.Stop()
}

func TestGetGroups(t *testing.T) {
	// logrus.Infof("---TestGetGroups---")
	as := auth.NewAclStoreFile(aclFile)
	err := as.Open()
	assert.NoError(t, err)
	groups := as.GetGroups(testDevice1)
	assert.GreaterOrEqual(t, len(groups), 1)
	as.Close()
}
func TestGetRole(t *testing.T) {
	// logrus.Infof("---TestGetRole---")
	as := auth.NewAclStoreFile(aclFile)
	err := as.Open()
	assert.NoError(t, err)
	groups := as.GetGroups(testDevice1)
	role := as.GetRole(testUser1, groups)
	assert.Equal(t, auth.GroupRoleManager, role)
	as.Close()
}

func TestHasPermission(t *testing.T) {
	logrus.Infof("---TestHasPermission---")

	ah := auth.NewAuthHandler(auth.NewAclStoreFile(aclFile))
	ah.Start()
	hasPerm := ah.HasPermission(auth.GroupRoleThing, false, hubclient.MessageTypeTD)
	assert.True(t, hasPerm)
	hasPerm = ah.HasPermission(auth.GroupRoleEditor, false, hubclient.MessageTypeTD)
	assert.True(t, hasPerm)
	hasPerm = ah.HasPermission(auth.GroupRoleViewer, false, hubclient.MessageTypeTD)
	assert.True(t, hasPerm)
	hasPerm = ah.HasPermission(auth.GroupRoleManager, false, hubclient.MessageTypeTD)
	assert.True(t, hasPerm)
	hasPerm = ah.HasPermission(auth.GroupRoleNone, false, hubclient.MessageTypeTD)
	assert.False(t, hasPerm)
	ah.Stop()

}

func TestCheckDeviceAuthorization(t *testing.T) {
	logrus.Infof("---TestCheckAuthorization---")
	ah := auth.NewAuthHandler(auth.NewAclStoreFile(aclFile))
	ah.Start()

	userName := "pub1"
	thingID := "urn:zone1:pub1:device1:sensor1"
	writing := true
	msgType := hubclient.MessageTypeTD
	isAuthorized := ah.CheckAuthorization(userName, certsetup.OUIoTDevice, thingID, writing, msgType)
	assert.True(t, isAuthorized)

	isAuthorized = ah.CheckAuthorization(userName, "", thingID, writing, msgType)
	assert.False(t, isAuthorized)
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
