package capnpserver

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capserializer"
)

// ManageAuthnCapnpServer provides the capnp RPC server for authentication management
type ManageAuthnCapnpServer struct {
	svc authn.IManageAuthn
}

func (capsrv *ManageAuthnCapnpServer) AddUser(
	ctx context.Context, call hubapi.CapManageAuthn_addUser) error {

	args := call.Args()
	loginID, _ := args.LoginID()
	newPassword, _ := args.Password()
	password, err := capsrv.svc.AddUser(ctx, loginID, newPassword)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		_ = res.SetPassword(password)
	}
	return err
}

func (capsrv *ManageAuthnCapnpServer) ListUsers(
	ctx context.Context, call hubapi.CapManageAuthn_listUsers) error {

	profiles, err := capsrv.svc.ListUsers(ctx)
	if err == nil {
		resp, _ := call.AllocResults()
		profilesCapnp := capserializer.MarshalUserProfileList(profiles)
		err = resp.SetProfiles(profilesCapnp)
	}
	return err
}

func (capsrv *ManageAuthnCapnpServer) RemoveUser(
	ctx context.Context, call hubapi.CapManageAuthn_removeUser) error {
	args := call.Args()
	loginID, _ := args.LoginID()
	err := capsrv.svc.RemoveUser(ctx, loginID)
	return err
}

func (capsrv *ManageAuthnCapnpServer) ResetPassword(
	ctx context.Context, call hubapi.CapManageAuthn_resetPassword) error {
	args := call.Args()
	loginID, _ := args.LoginID()
	newPassword, _ := args.NewPassword()
	password, err := capsrv.svc.ResetPassword(ctx, loginID, newPassword)
	if err == nil {
		resp, _ := call.AllocResults()
		err = resp.SetPassword(password)
	}
	return err
}
