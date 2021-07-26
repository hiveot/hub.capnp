package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/sirupsen/logrus"
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
	aclStore  IAclStoreReader
	unpwStore IUnpwStoreReader // how to allow nil check?
	// unpwStore *PasswordFileStore
}

// CreatePasswordHash for the given password
// This just creates the hash and does not update the store. See also VerifyPasswordHash
//  password to ahsh
//  algo is the algorithm to use, PWHASH_ARGON2id or PWHASH_BCRYPT
//  iterations for argon2id, default is 10
func CreatePasswordHash(password string, algo string, iterations uint) (hash string, err error) {
	if algo == PWHASH_ARGON2id {
		if iterations <= 0 {
			iterations = 10
		}
		params := argon2id.DefaultParams
		params.Iterations = uint32(iterations)
		hash, err = argon2id.CreateHash(password, params)
	} else if algo == PWHASH_BCRYPT {
		var hashBytes []byte
		hashBytes, err = bcrypt.GenerateFromPassword([]byte(password), 0)
		hash = string(hashBytes)
	} else {
		err = fmt.Errorf("CreatePasswordHash: Unsupported hashing algorithm '%s'", algo)
	}
	return hash, err
}

// CheckLoginPassword verifies if the given password is valid for login
// Returns true if valid, false if the user is unknown or the password is invalid
func (ah *AuthHandler) CheckUsernamePassword(loginName string, password string) bool {
	// Todo: configure hashing method
	algo := PWHASH_ARGON2id
	h := ah.unpwStore.GetPasswordHash(loginName)
	match := ah.VerifyPasswordHash(h, password, algo)
	logrus.Infof("CheckUsernamePassword: loginName=%s, match=%v", loginName, match)
	return match
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
		if role != GroupRoleThing && role != GroupRoleEditor && role != GroupRoleManager {
			err = fmt.Errorf("HasPermission: Role %s has no write access to message type %s", role, messageType)
		}
	} else {
		if role != GroupRoleThing && role != GroupRoleEditor && role != GroupRoleManager && role != GroupRoleViewer {
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
	logrus.Infof("AuthHandler.Start Opening ACL store")
	if ah.aclStore == nil || ah.unpwStore == nil {
		return fmt.Errorf("AuthHandler.Start: Missing ACL or password store")
	}
	err := ah.aclStore.Open()
	if err != nil {
		return err
	}
	logrus.Infof("AuthHandler.Start Opening password store")
	err = ah.unpwStore.Open()
	if err != nil {
		logrus.Errorf("AuthHandler.Start Failed opening password store: %s", err)
		logrus.Panic()
		return err
	}
	logrus.Infof("AuthHandler.Start Success")

	return nil
}

// Stop the auth handler and close the ACL and password store access.
func (ah *AuthHandler) Stop() {
	if ah.aclStore != nil {
		ah.aclStore.Close()
	}
	if ah.unpwStore != nil {
		ah.unpwStore.Close()
	}
}

// VerifyPasswordHash verifies if the given hash matches the password
// This does not access the store
//  hash to verify
//  password to verify against
//  algo is the algorithm to use, PWHASH_ARGON2id or PWHASH_BCRYPT
// returns true on success, or false on mismatch
func (ah *AuthHandler) VerifyPasswordHash(hash string, password string, algo string) bool {
	if algo == PWHASH_ARGON2id {
		match, _ := argon2id.ComparePasswordAndHash(password, hash)
		return match
	} else if algo == PWHASH_BCRYPT {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		return (err == nil)
	}
	return false
}

// NewAuthHandler creates a new instance of the authentication/authorization handler for validation only.
//  aclStore provides the functions to read authorization rules
//  unpwStore provides the functions to read username password hashes. nil to disable
func NewAuthHandler(aclStore IAclStoreReader, unpwStore IUnpwStoreReader) *AuthHandler {
	ah := AuthHandler{
		aclStore:  aclStore,
		unpwStore: unpwStore,
	}
	return &ah
}
