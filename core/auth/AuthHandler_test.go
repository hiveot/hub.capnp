package auth_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/core/auth"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
)

// TODO: table driven testing for all combinations

func TestStartStop(t *testing.T) {
	logrus.Infof("---TestStartStop---")
	ah := auth.NewAuthHandler()
	err := ah.Start()
	assert.NoError(t, err)
	ah.Stop()
}

func TestIsPublisher(t *testing.T) {
	logrus.Infof("---TestIsPublisher---")
	deviceID := "device1"
	thingID1 := "urn:zone:" + deviceID + ":sensor1:temperature"
	thingID2 := "urn:zone:" + deviceID + ":sensor1"
	thingID3 := "urn:zone:" + deviceID + ""
	ah := auth.NewAuthHandler()
	isPublisher := ah.IsPublisher(deviceID, thingID1)
	assert.True(t, isPublisher)
	isPublisher = ah.IsPublisher(deviceID, thingID2)
	assert.False(t, isPublisher)
	isPublisher = ah.IsPublisher(deviceID, thingID3)
	assert.False(t, isPublisher)
}
func TestHasPermission(t *testing.T) {
	logrus.Infof("---TestHasPermission---")

	ah := auth.NewAuthHandler()
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
}

func TestCheckDeviceAuthorization(t *testing.T) {
	logrus.Infof("---TestCheckAuthorization---")
	ah := auth.NewAuthHandler()
	userName := "pub1"
	thingID := "urn:zone1:pub1:device1:sensor1"
	writing := true
	msgType := hubclient.MessageTypeTD
	isAuthorized := ah.CheckAuthorization(userName, certsetup.OUIoTDevice, thingID, writing, msgType)
	assert.True(t, isAuthorized)

	isAuthorized = ah.CheckAuthorization(userName, "", thingID, writing, msgType)
	assert.False(t, isAuthorized)
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
