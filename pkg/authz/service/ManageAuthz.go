package service

import (
	"context"

	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/service/aclstore"
)

// ManageAuthz provides the capability to manage the authorization service
// This implements the IManageAuthorization interface
// TBD: add constraints if needed
type ManageAuthz struct {
	aclStore *aclstore.AclFileStore
	// the client requesting this session
	clientID string
}

// AddThing adds a Thing to a group
func (manageAuthz *ManageAuthz) AddThing(ctx context.Context, thingID string, groupName string) error {

	err := manageAuthz.aclStore.SetRole(ctx, thingID, groupName, authz.ClientRoleThing)
	return err
}

// Release this client capability. To be invoked after use has completed.
func (manageAuthz *ManageAuthz) Release() {
	// nothing to do here
}

// GetGroup returns the group with the given name, or an error if group is not found.
// GroupName must not be empty
func (manageAuthz *ManageAuthz) GetGroup(
	ctx context.Context, groupName string) (group authz.Group, err error) {

	group, err = manageAuthz.aclStore.GetGroup(ctx, groupName)
	return group, err
}

// GetGroupRoles returns a list of roles in groups the client is a member of.
func (manageAuthz *ManageAuthz) GetGroupRoles(
	ctx context.Context, clientID string) (roles authz.RoleMap, err error) {

	// simple pass through
	roles = manageAuthz.aclStore.GetGroupRoles(ctx, clientID)
	return roles, nil
}

// ListGroups returns the list of known groups
func (manageAuthz *ManageAuthz) ListGroups(
	ctx context.Context, limit int, offset int) (groups []authz.Group, err error) {

	groups = manageAuthz.aclStore.ListGroups(ctx, limit, offset)
	return groups, nil
}

// RemoveAll from all groups
func (mngAuthz *ManageAuthz) RemoveAll(ctx context.Context, clientID string) error {
	err := mngAuthz.aclStore.RemoveAll(ctx, clientID)
	return err
}

// RemoveClient from a group
func (mngAuthz *ManageAuthz) RemoveClient(ctx context.Context, clientID string, groupName string) error {
	err := mngAuthz.aclStore.Remove(ctx, clientID, groupName)
	return err
}

// RemoveThing removes a Thing from a group
func (manageAuthz *ManageAuthz) RemoveThing(ctx context.Context, thingID string, groupName string) error {

	err := manageAuthz.aclStore.Remove(ctx, thingID, groupName)
	return err
}

// SetClientRole sets the role for the client in a group
func (manageAuthz *ManageAuthz) SetClientRole(ctx context.Context, clientID string, groupName string, role string) error {
	err := manageAuthz.aclStore.SetRole(ctx, clientID, groupName, role)
	return err
}

// NewManageAuthz returns a new instance with the capability to manage the authorization service
func NewManageAuthz(aclStore *aclstore.AclFileStore, clientID string) *ManageAuthz {
	mngAuthz := &ManageAuthz{
		aclStore: aclStore,
		clientID: clientID,
	}
	return mngAuthz
}
