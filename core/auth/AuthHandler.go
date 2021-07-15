package auth

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/td"
)

// AuthHandler handlers client authorization for access to Things
// Work in progress. Currently authorizes IoT device access using certificates
type AuthHandler struct {
	aclStore IAclStoreReader
}

// CheckAuthorization tests if the client has access to the device for the given operation
// certOU, if provided gives full access to plugins and administrators
//
//  userName the login name or device ID of an authenticated client: consumer, plugin or device
//  certOU the OU of the user certificate if present.
//  thingID is the ID of the Thing to access
//  writing: true for writing to Thing, false for reading
//  messageType is one of: td, configure, event, action
// This returns false if access is denied or nil if allowed
func (auth *AuthHandler) CheckAuthorization(
	userName string, certOU string, thingID string, writing bool, messageType string) bool {

	if certOU == certsetup.OUPlugin || certOU == certsetup.OUAdmin {
		// plugins and admins have full permission
		return true
	} else if certOU == certsetup.OUIoTDevice {
		if !auth.IsPublisher(userName, thingID) {
			// publishers of IoT devices can not access devices of other publishers
			logrus.Infof("CheckAuthorization: Refused access by device '%s' to thingID '%s'. Thing belongs to a different publisher", userName, thingID)
			return false
		}
		return true
	}
	// anything else is allowed access if they are in the same group as the thing
	groups := auth.aclStore.GetGroups(thingID)
	role := auth.aclStore.GetRole(userName, groups)
	hasPerm := auth.HasPermission(role, writing, messageType)
	return hasPerm
}

// Determine if the consumer role allows the read/write operation
func (auth *AuthHandler) HasPermission(role string, writing bool, messageType string) bool {
	hasPermission := false
	// TODO: include message type
	if writing {
		hasPermission = (role == GroupRoleEditor || role == GroupRoleManager || role == GroupRoleThing)
	} else {
		hasPermission = (role == GroupRoleEditor || role == GroupRoleManager || role == GroupRoleViewer || role == GroupRoleThing)
	}
	return hasPermission
}

// IsPublisher checks if the deviceID is the publisher component of the thingID
// This is based on the predefined thingID format publisher:sensorID
func (auth *AuthHandler) IsPublisher(deviceID string, thingID string) bool {
	zone, publisherID, thingDeviceID, deviceType := td.SplitThingID(thingID)
	_ = zone
	_ = thingDeviceID
	_ = deviceType
	if publisherID != deviceID {
		return false
	}
	// permission granted: the publisher of the thingID is the device that is connected
	return true
}

// Start the authhandler. This loads its configuration and initializes its in-memory cache
func (auth *AuthHandler) Start() error {
	err := auth.aclStore.Open()
	return err
}

// Stop the auth handler.
func (auth *AuthHandler) Stop() {
	auth.aclStore.Close()
}

// NewAuthHandler creates a new instance of the authentication handler
func NewAuthHandler(aclStore IAclStoreReader) *AuthHandler {
	ah := AuthHandler{
		aclStore: aclStore,
	}
	return &ah
}
