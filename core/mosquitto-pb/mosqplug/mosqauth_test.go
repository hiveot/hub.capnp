package mosqplug_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/core/mosquitto-pb/mosqplug"
)

func TestAuthPluginInit(t *testing.T) {
	logrus.Infof("---TestAuthPluginInit---")
	mosqplug.AuthPluginInit(nil, nil, 0)
}

func TestAuthUnpwdCheck(t *testing.T) {
	logrus.Infof("---TestAuthUnpwdCheck---")
	username := "user1"
	password := "password1"
	clientID := "clientID1"
	clientIP := "ip"
	mosqplug.AuthUnpwdCheck(clientID, username, password, clientIP)
}

func TestAuthAclCheck(t *testing.T) {
	logrus.Infof("---TestAuthAclCheck---")
	clientID := "clientID1"
	username := "user1"
	topic := "things/thingid1/td"
	access := mosqplug.MOSQ_ACL_SUBSCRIBE
	mosqplug.AuthAclCheck(clientID, username, topic, access, true)
}

func TestAuthPluginCleanup(t *testing.T) {
	logrus.Infof("---TestAuthPluginCleanup---")
	mosqplug.AuthPluginCleanup()
}
