package auth

import "github.com/wostzone/wostlib-go/pkg/certsetup"

// AuthHandler handlers client authorization for access to Things
type AuthHandler struct {
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
func (auth *AuthHandler) CheckAuthorization(userName string, certOU string, thingID string, writing bool, messageType string) bool {
	if certOU == certsetup.OUPlugin || certOU == certsetup.OUAdmin {
		return true
	}

	// devices can do anything on with their own Things
	if auth.IsPublisher(userName, thingID) {
		return true
	}

	groups := auth.GetGroups(thingID)
	role := auth.GetRole(userName, groups)
	hasPerm := auth.HasPermission(role, writing, messageType)
	return hasPerm
}

// Return the group IDs of the groups the thing is a member of
func (auth *AuthHandler) GetGroups(thingID string) []string {
	// FIXME: make this work
	return []string{}
}

// Return the highest role the client has in a group
func (auth *AuthHandler) GetRole(clientID string, groups []string) string {
	// FIXME: make this work
	return GroupRoleManager
}

// Determine if the role allows the operation
func (auth *AuthHandler) HasPermission(role string, writing bool, messageType string) bool {
	// FIXME: make this work
	return true
}
func (auth *AuthHandler) IsPublisher(clientID string, thingID string) bool {
	// FIXME: make this work
	return false
}

// Start the authhandler. This loads its configuration and initializes its in-memory cache
func (auth *AuthHandler) Start() {
}

// Stop the auth handler.
func (auth *AuthHandler) Stop() {
}

// NewAuthHandler creates a new instance of the authentication handler
func NewAuthHandler() *AuthHandler {
	ah := AuthHandler{}
	return &ah
}
