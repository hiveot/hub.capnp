package capnpclient

import (
	"context"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capserializer"
)

// UserAuthnCapnpClient provides the POGS wrapper around the capnp user API
// This implements the IUserAuthn interface
type UserAuthnCapnpClient struct {
	connection *rpc.Conn           // connection to capnp server
	capability hubapi.CapUserAuthn // capnp client of the user profile service
}

// GetProfile returns the user's profile.
// Login or Refresh must have been successfully called first.
func (cl *UserAuthnCapnpClient) GetProfile(ctx context.Context) (profile authn.UserProfile, err error) {
	method, release := cl.capability.GetProfile(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		profileCapnp, _ := resp.Profile()
		profile = capserializer.UnmarshalUserProfile(profileCapnp)
	}
	return profile, err
}

func (cl *UserAuthnCapnpClient) Login(
	ctx context.Context, password string) (authToken, refreshToken string, err error) {

	method, release := cl.capability.Login(ctx, func(params hubapi.CapUserAuthn_login_Params) error {
		err2 := params.SetPassword(password)
		return err2
	})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		authToken, _ = resp.AuthToken()
		refreshToken, _ = resp.RefreshToken()
	}
	return authToken, refreshToken, err
}

func (cl *UserAuthnCapnpClient) Logout(
	ctx context.Context, refreshToken string) (err error) {

	method, release := cl.capability.Logout(ctx, func(params hubapi.CapUserAuthn_logout_Params) error {
		err2 := params.SetRefreshToken(refreshToken)
		return err2
	})
	defer release()
	_, err = method.Struct()
	return err
}

func (cl *UserAuthnCapnpClient) Refresh(
	ctx context.Context, refreshToken string) (newAuthToken, newRefreshToken string, err error) {

	method, release := cl.capability.Refresh(ctx, func(params hubapi.CapUserAuthn_refresh_Params) error {
		err2 := params.SetRefreshToken(refreshToken)
		return err2
	})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		newAuthToken, _ = resp.NewAuthToken()
		newRefreshToken, _ = resp.NewRefreshToken()
	}
	return newAuthToken, newRefreshToken, err
}

func (cl *UserAuthnCapnpClient) Release() {
	cl.capability.Release()
}

// SetPassword Login or Refresh must have been successfully called first.
func (cl *UserAuthnCapnpClient) SetPassword(ctx context.Context, newPassword string) (err error) {
	method, release := cl.capability.SetPassword(ctx, func(params hubapi.CapUserAuthn_setPassword_Params) error {
		err2 := params.SetNewPassword(newPassword)
		return err2
	})
	defer release()
	_, err = method.Struct()
	return err
}

// SetProfile on the server
// Login or Refresh must have been successfully called first.
func (cl *UserAuthnCapnpClient) SetProfile(ctx context.Context, newProfile authn.UserProfile) (err error) {
	method, release := cl.capability.SetProfile(ctx,
		func(params hubapi.CapUserAuthn_setProfile_Params) error {
			profileCapnp := capserializer.MarshalUserProfile(newProfile)
			err2 := params.SetNewProfile(profileCapnp)
			return err2
		})
	defer release()
	_, err = method.Struct()
	return err
}

// NewUserAuthnCapnpClient returns a capnp client instance of the capability
func NewUserAuthnCapnpClient(capability hubapi.CapUserAuthn) *UserAuthnCapnpClient {
	cl := &UserAuthnCapnpClient{
		capability: capability,
	}
	return cl
}
