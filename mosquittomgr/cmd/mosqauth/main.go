// package main with the mosquitto auth plugin
package main

//#cgo CFLAGS: -g  -fPIC -I/usr/local/include -I./
//#cgo LDFLAGS: -L. -shared

import "C"
import (
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/hub/authz/pkg/aclstore"
	"github.com/wostzone/hub/authz/pkg/authorize"
	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost-go/pkg/consumedthing"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/tlsserver"
)

// Constants from mosquitto.h
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

// Default filenames for auth and logging
const (
	DefaultLogFile  = "authplug.log"
	DefaultLogLevel = "warning"
)

// Configuration keys using auth_opt_xxx in mosquitto.conf
const (
	MosqOptLogFile        = "logFile"
	MosqOptLogLevel       = "logLevel"
	MosqOptAclFile        = "aclFile"
	MosqOptServerCertFile = "serverCertFile"
)

var jwtAuthenticator *tlsserver.JWTAuthenticator
var authorizer *authorize.Authorizer

//var homeFolder string
//var hubConfig *config.HubConfig

// MosqAuthConfig is a Mosquitto authentication plugin configuration
// Authentication is handled by the auth module that serves not just mosquitto
// but also other services such as provisioning, the directory, and other services.
type MosqAuthConfig struct {
}

//export AuthPluginInit
func AuthPluginInit(keys []string, values []string, authOptsNum int) {

	logFile := DefaultLogFile
	logLevel := DefaultLogLevel
	aclFile := aclstore.DefaultAclFile
	serverCertFile := ""
	for index, key := range keys {
		if key == MosqOptLogFile {
			logFile = values[index]
		} else if key == MosqOptLogLevel {
			logLevel = values[index]
		} else if key == MosqOptAclFile {
			aclFile = values[index]
		} else if key == MosqOptServerCertFile {
			serverCertFile = values[index]
		}
	}
	logging.SetLogging(logLevel, logFile)
	logrus.Warningf("mosqauth: AuthPluginInit invoked. Keys=%s", keys)

	// The file based store is the only option for now
	if aclFile == "" {
		aclFile = aclstore.DefaultAclFile
	}
	aclStore := aclstore.NewAclFileStore(aclFile, "mosqauth.AuthPluginInit")
	authorizer = authorize.NewAuthorizer(aclStore)
	err := authorizer.Start()
	if err != nil {
		logrus.Errorf("Authorizer failed to start: %s", err)
	}

	// Tokens are signed by the server private key.
	// The server certificate holds the public key for verifying JWT access tokens.
	if serverCertFile != "" {
		serverCert, err := certsclient.LoadX509CertFromPEM(serverCertFile)
		if err != nil {
			logrus.Warningf("Failed loading the server certificate for JWT verification from '%s': %s", serverCertFile, err)
		} else {
			serverKey := certsclient.PublicKeyFromCert(serverCert)
			jwtAuthenticator = tlsserver.NewJWTAuthenticator(serverKey)
		}
	}
}

// AuthUnpwdCheck checks for a correct username/password (jwt token)
// This verifies the JWT access token given as the password
//  clientID used to connect
//  username is the login user name
//  password is the JWT access token
//  clientIP
//  certSubjName when authenticated using a certificate instead of username/password
// Returns:
//  MOSQ_ERR_SUCCESS if the user is authenticated
//  MOSQ_ERR_PLUGIN_DEFER if we do not wish to handle this check
//export AuthUnpwdCheck
func AuthUnpwdCheck(clientID string, username string, password string, clientIP string, certSubjName string) uint8 {

	logrus.Warningf("AuthUnpwdCheck (JWT token check): clientID=%s, username=%s, clientIP=%s, subjname=%s",
		clientID, username, clientIP, certSubjName)

	// any client certificate is a match
	match := certSubjName != ""
	if !match {
		//match = authenticator.VerifyUsernamePassword(username, password)
		if !match && jwtAuthenticator != nil {
			jwtToken, claims, err := jwtAuthenticator.DecodeToken(password)
			_ = jwtToken
			if err != nil {
				logrus.Warningf("AuthUnpwdCheck: Invalid JWT token '%s' for user %s: %s", password, username, err)
				match = false
			} else if claims.Username != username {
				logrus.Warningf("AuthUnpwdCheck: User '%s' attempt to login with token that belongs to user '%s'", username, claims.Username)
				match = false
			} else {
				logrus.Infof("AuthUnpwdCheck: User '%s' authenticated with a valid JWT token", username)
				match = true
			}
		} else {
			logrus.Warningf("AuthUnpwdCheck: No JWT Authenticator. User login of '%s' declined", username)
		}
	}
	if !match {
		return MOSQ_ERR_PLUGIN_DEFER
	}
	return MOSQ_ERR_SUCCESS
}

// AuthAclCheck checks if the user has access to the topic
// If certificate authentication was used the certSubjName includes the OU of the client.
// The authorizer engine can decide to give extra access to clients based on their OU.
//
// This:
//   1. determines the thingID to access from the topic
//   2. determine the groups the Thing is in
//   3. determine the highest permission of the user if a member of one of those groups
//
//  clientID used to connect to the message bus
//  userID login ID of user when logging in with username/password
//  certSubjName: certificate subject when client certificate authentication is used
//  topic to validate
//  access: MOSQ_ACL_SUBSCRIBE, MOSQ_ACL_READ, MOSQ_ACL_WRITE
//
// returns
//  MOSQ_ERR_ACL_DENIED if access was not granted
//  MOSQ_ERR_UNKNOWN for an application specific error
//  MOSQ_ERR_SUCCESS if access is granted
//  MOSQ_ERR_PLUGIN_DEFER if we do not wish to handle this check
//export AuthAclCheck
func AuthAclCheck(clientID string, userID string, certSubjName string, topic string, access int) uint8 {
	var certOU = ""

	// what OU does this client belong to?
	parts := strings.Split(certSubjName, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "OU=") {
			certOU = part[3:]
		}
	}

	// topic format: things/{thingID}/td|property|event|action
	parts = strings.Split(topic, "/")
	if len(parts) < 3 {
		logrus.Infof("mosqauth: AuthAclCheck Invalid topic format '%s'. Expected min 3 parts.", topic)
		return MOSQ_ERR_ACL_DENIED
	}
	thingID := parts[1]
	messageType := parts[2]
	writing := access == MOSQ_ACL_WRITE
	authType := authorize.AuthRead
	if writing {
		switch messageType {
		case consumedthing.TopicTypeTD:
			authType = authorize.AuthPubTD
		case consumedthing.TopicTypeAction:
			authType = authorize.AuthEmitAction
		case consumedthing.TopicTypeEvent:
			authType = authorize.AuthPubEvent // including property value
		// how to determine if an action is a write property?
		// invoking an action currently allows for writing configuration properties
		//case mqttbinding.TopicTypeWrite:
		//	authType = authorize.AuthWriteProperty
		default:
			logrus.Warningf("mosqauth. Unknown message type in topic: %s from client %s / user %s", topic, clientID, userID)
			authType = messageType
		}
	}
	authorized := authorizer.VerifyAuthorization(userID, certOU, thingID, authType)
	if !authorized {
		logrus.Warningf("mosqauth: AuthAclCheck Access DENIED: clientID=%s, username=%s, certOU=%s, topic=%s, access=%d",
			clientID, userID, certOU, topic, access)
		return MOSQ_ERR_ACL_DENIED
	}

	logrus.Infof("mosqauth: AuthAclCheck Access Granted: clientID=%s, username=%s, certOU=%s, topic=%s, access=%d",
		clientID, userID, certOU, topic, access)
	return MOSQ_ERR_SUCCESS
	// return
}

//export AuthPluginCleanup
func AuthPluginCleanup() {
	logrus.Info("AuthPluginCleanup: Cleaning up plugin")
	if authorizer != nil {
		authorizer.Stop()
		authorizer = nil
	}
}

func main() {}
