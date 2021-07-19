package authhandler

import (
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/wostzone/hub/pkg/auth"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/td"
	"golang.org/x/crypto/bcrypt"
)

// supported password hashes
const (
	PWHASH_ARGON2id = "argon2id"
	PWHASH_BCRYPT   = "bcrypt" // fallback in case argon2i cannot be used
)

// AuthHandler handlers client authentication and authorization for access to Things
//  Authentication applies only to username/password logins. Certificate authentication is handled by
// TLS itself when using client certificas. See idprov for issuing and renewing client certificates.
//  Authorization uses access control lists with group membership and roles to determine if a client
// is authorized to receive or post a message. This applies to all users of the message bus,
// regardless of how they are authenticated.
type AuthHandler struct {
	aclStore  auth.IAclStoreReader
	unpwStore auth.IUnpwStoreReader
}

// CheckLoginPassword verifies if the given password is valid for login
// Returns true if valid, false if the user is unknown or the password is invalid
func (ah *AuthHandler) CheckUsernamePassword(loginName string, password string) error {
	var err error

	// Todo: configure hashing method
	pwhash := PWHASH_ARGON2id
	h := ah.unpwStore.GetPasswordHash(loginName)
	if h == "" {
		// this is not a valid password, use to reduce timing difference with valid user
		// TODO: iterations and memory must match the configured encoding
		h = "$argon2i$v=19$m=4096,t=3,p=1$dGhpc2lzbXlzYWx0$WzR/Vji668772vv++KMoKlaN3AJA1BGR7bCGt4Q2fsA"
	}

	if pwhash == PWHASH_ARGON2id {
		match, err2 := argon2id.ComparePasswordAndHash(password, h)
		if err2 != nil {
			err = err2
		} else if !match {
			return fmt.Errorf("CheckUsernamePassword: Invalid username or password")
		}
	} else if pwhash == PWHASH_BCRYPT {
		err = bcrypt.CompareHashAndPassword([]byte(h), []byte(password))
	} else {
		err = fmt.Errorf("CheckUsernamePassword: Unsupported password hash '%s'", pwhash)
	}

	return err
}

// CheckAuthorization tests if the client has access to the device for the given operation
// certOU, if provided gives full access to plugins and administrators
//
//  username is the login name or device ID of the client seeking permission
//  certOU is the OU of the client seeking permission, if client certificate if used to authenticate. "" to ignore.
//  thingID is the ID of the Thing to access
//  writing: true for writing to Thing, false for reading
//  messageType is one of: td, configure, event, action
// This returns error if access is denied or nil if allowed
func (ah *AuthHandler) CheckAuthorization(
	username string, certOU string, thingID string, writing bool, messageType string) error {

	// plugins and admins have full permission
	if certOU == certsetup.OUPlugin || certOU == certsetup.OUAdmin {
		return nil
	} else if certOU == certsetup.OUIoTDevice {
		// A publishing device of IoT things can access their own things
		if !ah.IsPublisher(username, thingID) {
			err := fmt.Errorf("CheckAuthorization: Refused access by device '%s' to thingID '%s'. Thing belongs to a different publisher", username, thingID)
			return err
		}
		return nil
	}
	// Consumers must be in the same group as the thing
	groups := ah.aclStore.GetGroups(thingID)
	if len(groups) == 0 {
		err := fmt.Errorf("CheckAuthorization: User '%s' not in same group as thingID '%s'", username, thingID)
		return err
	}
	// Consumer must have the correct read/write role for the message type (td, action, ..)
	role := ah.aclStore.GetRole(username, groups)
	err := ah.HasPermission(role, writing, messageType)
	return err
}

// Determine if the consumer role allows the read/write operation
// * The viewer role only has read access
// * The editor role has read/write access to thing values
// * The manager role has read/write access to thing configuration and values
// * The thing role has read/write access (by its own device only - separate check)
// Returns an error if permission is denied, nil if granted
func (ah *AuthHandler) HasPermission(role string, writing bool, messageType string) error {
	var err error
	// TODO: include message type
	if writing {
		if role != auth.GroupRoleThing && role != auth.GroupRoleEditor && role != auth.GroupRoleManager {
			err = fmt.Errorf("HasPermission: Role %s has no write access to message type %s", role, messageType)
		}
	} else {
		if role != auth.GroupRoleThing && role != auth.GroupRoleEditor && role != auth.GroupRoleManager && role != auth.GroupRoleViewer {
			err = fmt.Errorf("HasPermission: Role %s has no read access to message type %s", role, messageType)
		}
	}
	return err
}

// IsPublisher checks if the deviceID is the publisher of the thingID.
// This requires that the thingID is formatted as "urn:publisherID:sensorID...""
// Returns true if the deviceID is the publisher of the thingID, false if not.
func (ah *AuthHandler) IsPublisher(deviceID string, thingID string) bool {
	zone, publisherID, thingDeviceID, deviceType := td.SplitThingID(thingID)
	_ = zone
	_ = thingDeviceID
	_ = deviceType
	return publisherID == deviceID
}

// Start the authhandler. This opens the ACL and password stores for reading
func (ah *AuthHandler) Start() error {
	err := ah.aclStore.Open()
	if err != nil {
		return err
	}
	err = ah.unpwStore.Open()
	return err
}

// Stop the auth handler and close the ACL and password store access.
func (ah *AuthHandler) Stop() {
	ah.aclStore.Close()
	ah.unpwStore.Close()
}

// NewAuthHandler creates a new instance of the authentication/authorization handler for validation only.
//  aclStore provides the functions to read authorization rules
//  unpwStore provides the functions to read username password hashes
func NewAuthHandler(aclStore auth.IAclStoreReader, unpwStore auth.IUnpwStoreReader) *AuthHandler {
	ah := AuthHandler{
		aclStore:  aclStore,
		unpwStore: unpwStore,
	}
	return &ah
}
