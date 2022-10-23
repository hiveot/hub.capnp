package capnpserver

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/capnp4POGS"
)

// ManageAuthzCapnpServer provides the capnp RPC server for Client authorization
type ManageAuthzCapnpServer struct {
	srv authz.IManageAuthz
}

func (capsrv *ManageAuthzCapnpServer) AddThing(
	ctx context.Context, call hubapi.CapManageAuthz_addThing) (err error) {

	args := call.Args()
	thingID, _ := args.ThingID()
	groupName, _ := args.GroupName()
	err = capsrv.srv.AddThing(ctx, thingID, groupName)
	return err
}

func (capsrv *ManageAuthzCapnpServer) GetGroup(
	ctx context.Context, call hubapi.CapManageAuthz_getGroup) (err error) {

	args := call.Args()
	groupName, _ := args.GroupName()
	grp, err := capsrv.srv.GetGroup(ctx, groupName)
	if err == nil {
		grpCap := capnp4POGS.GroupPOGS2Capnp(grp)
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
		roleMapCap := capnp4POGS.RoleMapPOGS2Capnp(roleMap)
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
		grpListCap := capnp4POGS.GroupListPOGS2Capnp(grpList)
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
	thingID, _ := args.ThingID()
	groupName, _ := args.GroupName()
	err = capsrv.srv.RemoveClient(ctx, thingID, groupName)
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
