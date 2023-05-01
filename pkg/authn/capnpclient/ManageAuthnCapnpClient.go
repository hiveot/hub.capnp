package capnpclient

import (
	"context"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capserializer"
)

// ManageAuthnCapnpClient provides the POGS wrapper around the capnp user API
// This implements the IManageAuthn interface
type ManageAuthnCapnpClient struct {
	connection *rpc.Conn             // connection to capnp server
	capability hubapi.CapManageAuthn // capnp client of the user profile service
}

func (cl *ManageAuthnCapnpClient) AddUser(ctx context.Context,
	loginID string, newPassword string) (password string, err error) {

	method, release := cl.capability.AddUser(ctx, func(params hubapi.CapManageAuthn_addUser_Params) error {
		err2 := params.SetLoginID(loginID)
		_ = params.SetPassword(newPassword)
		return err2
	})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		password, err = resp.Password()
	}
	return password, err
}

// ListUsers provide a list of users and their info
func (cl *ManageAuthnCapnpClient) ListUsers(
	ctx context.Context) (profiles []authn.UserProfile, err error) {

	method, release := cl.capability.ListUsers(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		profileListCapnp, err2 := resp.Profiles()
		err = err2
		profiles = capserializer.UnmarshalUserProfileList(profileListCapnp)
	}
	return profiles, err
}

// RemoveUser removes a user and disables login
// Existing tokens are immediately expired (tbd)
func (cl *ManageAuthnCapnpClient) RemoveUser(ctx context.Context, loginID string) (err error) {

	method, release := cl.capability.RemoveUser(ctx, func(params hubapi.CapManageAuthn_removeUser_Params) error {
		err2 := params.SetLoginID(loginID)
		return err2
	})
	defer release()
	_, err = method.Struct()
	return err
}

// ResetPassword reset the user's password and returns a new password
func (cl *ManageAuthnCapnpClient) ResetPassword(
	ctx context.Context, loginID string, newPassword string) (password string, err error) {

	method, release := cl.capability.ResetPassword(ctx, func(params hubapi.CapManageAuthn_resetPassword_Params) error {
		err2 := params.SetLoginID(loginID)
		params.SetNewPassword(newPassword)
		return err2
	})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		password, err = resp.Password()
	}
	return password, err
}

func (cl *ManageAuthnCapnpClient) Release() {
	cl.capability.Release()
}

func NewManageAuthnCapnpClient(capability hubapi.CapManageAuthn) *ManageAuthnCapnpClient {
	cl := &ManageAuthnCapnpClient{
		capability: capability,
	}
	return cl
}
