package capnpserver

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/capserializer"
)

// ManageAuthzCapnpServer provides the capnp RPC server for Client authorization
type ManageAuthzCapnpServer struct {
	srv authz.IManageAuthz
}

func (capsrv *ManageAuthzCapnpServer) AddThing(
	ctx context.Context, call hubapi.CapManageAuthz_addThing) (err error) {

	args := call.Args()
	thingAddr, _ := args.ThingAddr()
	groupName, _ := args.GroupName()
	err = capsrv.srv.AddThing(ctx, thingAddr, groupName)
	return err
}

func (capsrv *ManageAuthzCapnpServer) GetGroup(
	ctx context.Context, call hubapi.CapManageAuthz_getGroup) (err error) {

	args := call.Args()
	groupName, _ := args.GroupName()
	grp, err := capsrv.srv.GetGroup(ctx, groupName)
	if err == nil {
		grpCap := capserializer.MarshalGroup(grp)
		res, _ := call.AllocResults()
		err = res.SetGroup(grpCap)
	} else {
		logrus.Infof("group '%s' does not exist", groupName)
	}
	return err
}

func (capsrv *ManageAuthzCapnpServer) GetGroupRoles(
	ctx context.Context, call hubapi.CapManageAuthz_getGroupRoles) (err error) {

	args := call.Args()
	clientID, _ := args.ClientID()
	roleMap, err := capsrv.srv.GetGroupRoles(ctx, clientID)
	if err == nil {
		roleMapCap := capserializer.MarshalRoleMap(roleMap)
		res, _ := call.AllocResults()
		err = res.SetRoles(roleMapCap)
	}
	return err
}

func (capsrv *ManageAuthzCapnpServer) ListGroups(
	ctx context.Context, call hubapi.CapManageAuthz_listGroups) (err error) {

	args := call.Args()
	limit := args.Limit()
	offset := args.Offset()
	grpList, err := capsrv.srv.ListGroups(ctx, int(limit), int(offset))
	if err == nil {
		grpListCap := capserializer.MarshalGroupList(grpList)
		res, _ := call.AllocResults()
		err = res.SetGroups(grpListCap)
	}
	return err
}

func (capsrv *ManageAuthzCapnpServer) RemoveAll(
	ctx context.Context, call hubapi.CapManageAuthz_removeAll) (err error) {

	args := call.Args()
	clientID, _ := args.ClientID()
	err = capsrv.srv.RemoveAll(ctx, clientID)
	return err
}

func (capsrv *ManageAuthzCapnpServer) RemoveClient(
	ctx context.Context, call hubapi.CapManageAuthz_removeClient) (err error) {

	args := call.Args()
	clientID, _ := args.ClientID()
	groupName, _ := args.GroupName()
	err = capsrv.srv.RemoveClient(ctx, clientID, groupName)
	return err
}

func (capsrv *ManageAuthzCapnpServer) RemoveThing(
	ctx context.Context, call hubapi.CapManageAuthz_removeThing) (err error) {

	args := call.Args()
	thingAddr, _ := args.ThingAddr()
	groupName, _ := args.GroupName()
	err = capsrv.srv.RemoveClient(ctx, thingAddr, groupName)
	return err
}

func (capsrv *ManageAuthzCapnpServer) SetClientRole(
	ctx context.Context, call hubapi.CapManageAuthz_setClientRole) (err error) {

	args := call.Args()
	clientID, _ := args.ClientID()
	groupName, _ := args.GroupName()
	role, _ := args.Role()
	err = capsrv.srv.SetClientRole(ctx, clientID, groupName, role)
	return err
}
