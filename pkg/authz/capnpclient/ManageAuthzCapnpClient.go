package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/capserializer"
)

// ManageAuthzCapnpClient implements the IManageAuthorization interface client side
type ManageAuthzCapnpClient struct {
	capability hubapi.CapManageAuthz // capnp client of the authorization service
}

// AddThing adds a thing to a group
func (authz *ManageAuthzCapnpClient) AddThing(
	ctx context.Context, thingID, groupName string) (err error) {

	method, release := authz.capability.AddThing(ctx,
		func(params hubapi.CapManageAuthz_addThing_Params) error {
			err2 := params.SetThingID(thingID)
			_ = params.SetGroupName(groupName)
			return err2
		})
	defer release()

	_, err = method.Struct()
	return err
}

// GetGroup returns a list of roles for clients in the group
// GroupName must not be empty
func (authz *ManageAuthzCapnpClient) GetGroup(
	ctx context.Context, groupID string) (group authz.Group, err error) {

	method, release := authz.capability.GetGroup(ctx,
		func(params hubapi.CapManageAuthz_getGroup_Params) error {
			err := params.SetGroupName(groupID)
			return err
		})
	defer release()

	resp, err := method.Struct()
	if err == nil {
		capGroup, _ := resp.Group()

		group.Name, _ = capGroup.Name()
		rolesCapnp, _ := capGroup.MemberRoles()
		group.MemberRoles = capserializer.UnmarshalRoleMap(rolesCapnp)
	}
	return group, err
}

// GetGroupRoles returns a list of roles in groups the client is a member of.
func (authz *ManageAuthzCapnpClient) GetGroupRoles(
	ctx context.Context, clientID string) (roles authz.RoleMap, err error) {

	method, release := authz.capability.GetGroupRoles(ctx,
		func(params hubapi.CapManageAuthz_getGroupRoles_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()

	resp, err := method.Struct()
	if err == nil {
		rolesCapnp, _ := resp.Roles()
		roles = capserializer.UnmarshalRoleMap(rolesCapnp)
	}
	return roles, err
}

// ListGroups returns a list with available groups
func (authz *ManageAuthzCapnpClient) ListGroups(
	ctx context.Context, limit int, offset int) (groups []authz.Group, err error) {

	method, release := authz.capability.ListGroups(ctx,
		func(params hubapi.CapManageAuthz_listGroups_Params) error {
			params.SetLimit(int32(limit))
			params.SetOffset(int32(offset))
			return nil
		})
	defer release()

	resp, err := method.Struct()
	if err == nil {
		groupsCapnp, _ := resp.Groups()
		groups = capserializer.UnmarshalGroupList(groupsCapnp)
	}
	return groups, err
}

// RemoveAll removes a client from all groups
func (authz *ManageAuthzCapnpClient) RemoveAll(
	ctx context.Context, clientID string) (err error) {

	method, release := authz.capability.RemoveAll(ctx,
		func(params hubapi.CapManageAuthz_removeAll_Params) error {
			err2 := params.SetClientID(clientID)
			return err2
		})
	defer release()

	_, err = method.Struct()
	return err
}

// RemoveClient removes a client from a group
func (authz *ManageAuthzCapnpClient) RemoveClient(
	ctx context.Context, clientID string, groupName string) (err error) {

	method, release := authz.capability.RemoveClient(ctx,
		func(params hubapi.CapManageAuthz_removeClient_Params) error {
			err2 := params.SetClientID(clientID)
			_ = params.SetGroupName(groupName)
			return err2
		})
	defer release()

	_, err = method.Struct()
	return err
}

// RemoveThing removes a thing from a group
func (authz *ManageAuthzCapnpClient) RemoveThing(
	ctx context.Context, thingID string, groupName string) (err error) {

	method, release := authz.capability.RemoveThing(ctx,
		func(params hubapi.CapManageAuthz_removeThing_Params) error {
			err2 := params.SetThingID(thingID)
			_ = params.SetGroupName(groupName)
			return err2
		})
	defer release()

	_, err = method.Struct()
	return err
}

// SetClientRole sets the role for the client in a group.
// Note that 'things' are also clients. Things are added to groups with the role ClientRoleThing
//
// If the client is not a member of a group the client will be added.
// If the client is already a member of the group, its role will be replaced by the given role.
func (authz *ManageAuthzCapnpClient) SetClientRole(
	ctx context.Context, clientID string, groupName string, role string) error {

	method, release := authz.capability.SetClientRole(ctx,
		func(params hubapi.CapManageAuthz_setClientRole_Params) error {
			err2 := params.SetClientID(clientID)
			_ = params.SetGroupName(groupName)
			_ = params.SetRole(role)
			return err2
		})
	defer release()

	_, err := method.Struct()
	return err

}

// NewManageAuthzCapnpClient returns the capnp client for managing authz
func NewManageAuthzCapnpClient(cap hubapi.CapManageAuthz) *ManageAuthzCapnpClient {
	mngAuthz := &ManageAuthzCapnpClient{
		capability: cap,
	}
	return mngAuthz
}
