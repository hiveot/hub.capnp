package authorize

import (
	"fmt"
	"github.com/wostzone/hub/lib/client/pkg/certsclient"
	"github.com/wostzone/hub/lib/client/pkg/thing"

	"github.com/sirupsen/logrus"
)

// VerifyAuthorization defines the function to authorize access to a Thing.
// Intended for use by all Hub services that need authorization.
//  userID is the ID of the authenticated user as used in the group/rule list
//  certOU is the user's organization when using client certificates.
//  thingID is the ID of the Thing the user is trying to access
//  authType is message when writing, eg td.MessagetypeTD,... use MessagetypeNone for reading
type VerifyAuthorization func(userID string, certOU string,
	thingID string, authType string) bool

// AllGroupName is the name of the group that includes all things (no need to add things separately)
// Users that are a member of the all group will have access to all things based on their role.
const AllGroupName = "all"

// Authorizer handles client authorization for access to Things.
//
// Authorization uses access control lists with group membership and roles to determine if a client
// is authorized to receive or post a message. This applies to all users of the message bus,
// regardless of how they are authenticated.
type Authorizer struct {
	aclStore IAclStore
}

// VerifyAuthorization tests if the client has access to the device for the given operation
// The thingID is implicitly included in the 'all' group. Members of the 'all' group can access all things
// based on their role in that group.
//
//  username is the login name or device ID of the client seeking permission
//  certOU is the OU of the client seeking permission only if user is authenticate with a client certificate. "" to ignore. certsetup.OUPlugin for plugins
//  thingID is the ID of the Thing to access
//  authType is one of: AuthEmitAction|AuthPubEvent|AuthWriteProperty|AuthPubPropValue|AuthRead
//
// This returns false if access is denied or true if authorized
func (ah *Authorizer) VerifyAuthorization(
	userID string, certOU string, thingID string, authType string) bool {
	logrus.Debugf("CheckAuthorization: userID='%s' certOU='%s' thingID='%s', authType='%s'",
		userID, certOU, thingID, authType)

	// plugins and admins have full permission
	if certOU == certsclient.OUPlugin {
		return true
	} else if certOU == certsclient.OUIoTDevice {
		// A publishing device of IoT things can access their own things
		if !ah.IsPublisher(userID, thingID) {
			// err := fmt.Errorf("CheckAuthorization: Refused access by device '%s' to thingID '%s'. Thing belongs to a different publisher", username, thingID)
			logrus.Debugf("CheckAuthorization - IoT device cannot impersonate a different publisher: false")
			return false
		}
		logrus.Debugf("CheckAuthorization - IoT device can access its own devices: true")
		return true
	}
	// Consumers must be in the same group as the thing
	// all things are in the all group
	groups := ah.aclStore.GetGroups(thingID)
	groups = append(groups, AllGroupName)
	// if len(groups) == 0 {
	// 	err := fmt.Errorf("CheckAuthorization: User '%s' not in same group as thingID '%s'", username, thingID)
	// 	return err
	// }
	// Consumer must have the correct read/write role for the message type (td, action, ..)
	role := ah.aclStore.GetRole(userID, groups)
	allowed := ah.VerifyRolePermission(role, authType)
	logrus.Debugf("CheckAuthorization - role ('%s') in groups('%s') permission check for (authType=%s): %t", role, groups, authType, allowed)
	return allowed
}

// VerifyRolePermission determine if the consumer role allows the read/write operation
//  The viewer role has read access to properties
//  The operator role has access to invoke actions
//  The manager role has access to actions and read/write access to properties
//  The thing role has full read/write access to its own thing
//
// authType describes authorization to perform: EmitAction, PublishEvent, WriteProperty, "" to read
//
// Returns true if permission is denied, nil if granted
func (ah *Authorizer) VerifyRolePermission(role string, authType string) bool {

	if authType != AuthRead {
		// Things can write all their messages.
		// Check if the user is the thing itself first
		if role == GroupRoleThing {
			return true
		}
		// operators can control the thing
		if role == GroupRoleOperator && authType == AuthEmitAction {
			return true
		}
		// managers can configure and control the thing
		if role == GroupRoleManager && (authType == AuthWriteProperty || authType == AuthEmitAction) {
			return true
		}
		logrus.Debugf("VerifyRolePermission: Role %s has no write access to write type %s", role, authType)
	} else {
		// read access to all roles
		if role == GroupRoleThing || role == GroupRoleOperator || role == GroupRoleManager || role == GroupRoleViewer {
			return true
		}
		logrus.Debugf("VerifyRolePermission: Role %s has no read access", role)
	}
	return false
}

// IsPublisher checks if the deviceID is the publisher of the thingID.
// This requires that the thingID is formatted as "urn:zone:publisherID:sensorID...""
// Returns true if the deviceID is the publisher of the thingID, false if not.
func (ah *Authorizer) IsPublisher(deviceID string, thingID string) bool {
	zone, publisherID, thingDeviceID, deviceType := thing.SplitThingID(thingID)
	_ = zone
	_ = thingDeviceID
	_ = deviceType
	return publisherID == deviceID
}

// Start the authorizer. This opens the ACL store for reading
func (ah *Authorizer) Start() error {
	logrus.Infof("AuthHandler.Start Opening ACL store")
	if ah.aclStore == nil {
		err := fmt.Errorf("AuthHandler.Start: Missing ACL store")
		logrus.Errorf("%s", err)
		return err
	}
	err := ah.aclStore.Open()
	if err != nil {
		return err
	}
	logrus.Infof("AuthHandler.Start Success")
	return nil
}

// Stop the authn handler and close the ACL and password store access.
func (ah *Authorizer) Stop() {
	ah.aclStore.Close()
}

// NewAuthorizer creates a new instance of the authorization handler for managing authorization.
//  aclStore provides the functions to read and write authorization rules
func NewAuthorizer(aclStore IAclStore) *Authorizer {
	ah := Authorizer{
		aclStore: aclStore,
	}
	return &ah
}
