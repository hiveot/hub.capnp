// package main for both the protocol binding and the mosquitto auth plugin
package main

// #cgo CFLAGS: -g  -fPIC -I/usr/local/include -I./
// #cgo LDFLAGS: -L. -shared
import "C"
import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/core/auth"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// from mosquitto.h
const (
	MOSQ_ERR_AUTH_CONTINUE      = -4
	MOSQ_ERR_NO_SUBSCRIBERS     = -3
	MOSQ_ERR_SUB_EXISTS         = -2
	MOSQ_ERR_CONN_PENDING       = -1
	MOSQ_ERR_SUCCESS            = 0
	MOSQ_ERR_NOMEM              = 1
	MOSQ_ERR_PROTOCOL           = 2
	MOSQ_ERR_INVAL              = 3
	MOSQ_ERR_NO_CONN            = 4
	MOSQ_ERR_CONN_REFUSED       = 5
	MOSQ_ERR_NOT_FOUND          = 6
	MOSQ_ERR_CONN_LOST          = 7
	MOSQ_ERR_TLS                = 8
	MOSQ_ERR_PAYLOAD_SIZE       = 9
	MOSQ_ERR_NOT_SUPPORTED      = 10
	MOSQ_ERR_AUTH               = 11
	MOSQ_ERR_ACL_DENIED         = 12
	MOSQ_ERR_UNKNOWN            = 13
	MOSQ_ERR_ERRNO              = 14
	MOSQ_ERR_EAI                = 15
	MOSQ_ERR_PROXY              = 16
	MOSQ_ERR_PLUGIN_DEFER       = 17
	MOSQ_ERR_MALFORMED_UTF8     = 18
	MOSQ_ERR_KEEPALIVE          = 19
	MOSQ_ERR_LOOKUP             = 20
	MOSQ_ERR_MALFORMED_PACKET   = 21
	MOSQ_ERR_DUPLICATE_PROPERTY = 22
	MOSQ_ERR_TLS_HANDSHAKE      = 23
	MOSQ_ERR_QOS_NOT_SUPPORTED  = 24
	MOSQ_ERR_OVERSIZE_PACKET    = 25
	MOSQ_ERR_OCSP               = 26
)

// Autorization access requests
const (
	MOSQ_ACL_NONE      = 0x00
	MOSQ_ACL_READ      = 0x01 // check if client can read the topic, before it is sent to the client
	MOSQ_ACL_WRITE     = 0x02 // check if client can post to the topic, when it is received from the client
	MOSQ_ACL_SUBSCRIBE = 0x04 // check if client can subscribe to the topic (with wildcard)
)

var authHandler *auth.AuthHandler

//export AuthPluginInit
func AuthPluginInit(keys []string, values []string, authOptsNum int) {
	logrus.Warningf("mosqauth: AuthPluginInit invoked")

	// TODO: get level from config
	hubconfig.SetLogging("info", "")

	authHandler = auth.NewAuthHandler()
	authHandler.Start()
}

// AuthUnpwdCheck checks for a correct username/password
// This matches the given password against the stored password hash
// Returns:
//  MOSQ_ERR_AUTH if authentication failed
//  MOSQ_ERR_UNKNOWN for an application specific error
//  MOSQ_ERR_SUCCESS if the user is authenticated
//  MOSQ_ERR_PLUGIN_DEFER if we do not wish to handle this check
//export AuthUnpwdCheck
func AuthUnpwdCheck(clientID string, username string, password string, clientIP string) uint8 {
	// TODO: remove password logging
	logrus.Infof("mosqauth: AuthUnpwdCheck: clientID=%s, username=%s, pass=%s, clientIP=%s",
		clientID, username, password, clientIP)
	// TODO
	return MOSQ_ERR_PLUGIN_DEFER
}

// AuthAclCheck checks if the user has access to the topic
// This:
//   1. determines the thingID from the topic
//   2. determine the groups the thing is in
//   3. determine the highest permission of the user if a member of one of those groups
//
// TODO: currently this grants access.
//       This needs a group[thing,user/role] list loaded from the group configuration.
//
//  clientID
//  username
//  topic
//  access: MOSQ_ACL_SUBSCRIBE, MOSQ_ACL_READ, MOSQ_ACL_WRITE
//  certSubjName: certificate subject name "/L=/O=/OU=/CN=" or "" if no certificate was used
//
// returns
//  MOSQ_ERR_ACL_DENIED if access was not granted
//  MOSQ_ERR_UNKNOWN for an application specific error
//  MOSQ_ERR_SUCCESS if access is granted
//  MOSQ_ERR_PLUGIN_DEFER if we do not wish to handle this check
//export AuthAclCheck
func AuthAclCheck(clientID, userName, topic string, access int, certSubjName string) uint8 {
	logrus.Infof("mosqauth: AuthAclCheck clientID=%s, username=%s, topic=%s, access=%d, certSubj=%s",
		clientID, userName, topic, access, certSubjName)
	var certOU = ""

	// what OU does this client belong to?
	parts := strings.Split(certSubjName, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "OU=") {
			certOU = part[3:]
		}
	}

	// topic format: things/{publisherID}/{thingID}/td|configure|event|action|
	parts = strings.Split(topic, "/")
	if len(parts) < 4 {
		logrus.Infof("mosqauth: AuthAclCheck Invalid topic format '%s'. Expected min 4 parts.", topic)
		return MOSQ_ERR_ACL_DENIED
	}
	thingID := parts[2]
	messageType := parts[3]
	writing := (access == MOSQ_ACL_WRITE)
	hasPermission := authHandler.CheckAuthorization(userName, certOU, thingID, writing, messageType)
	if !hasPermission {
		return MOSQ_ERR_ACL_DENIED
	}

	return MOSQ_ERR_SUCCESS
	// return
}

//export AuthPluginCleanup
func AuthPluginCleanup() {
	logrus.Info("AuthPluginCleanup: Cleaning up plugin")
	if authHandler != nil {
		authHandler.Stop()
		authHandler = nil
	}
}

func main() {}

// Main entry to WoST protocol adapter for managing Mosquitto
// This setup the configuration from file and commandline parameters and launches the service
// func main() {
// 	svc := mosquittopb.NewMosquittoManager()
// 	hubConfig, err := hubconfig.LoadCommandlineConfig("", mosquittopb.PluginID, &svc.Config)
// 	if err != nil {
// 		logrus.Errorf("ERROR: Start aborted due to error")
// 		os.Exit(1)
// 	}

// 	err = svc.Start(hubConfig)
// 	if err != nil {
// 		logrus.Errorf("Logger: Failed to start: %s", err)
// 		os.Exit(1)
// 	}
// 	hubclient.WaitForSignal()
// 	svc.Stop()
// 	os.Exit(0)
// }
