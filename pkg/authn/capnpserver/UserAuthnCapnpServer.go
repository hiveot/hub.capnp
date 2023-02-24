package capnpserver

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capserializer"
)

// UserAuthnCapnpServer provides the capnp RPC server for client authentication
type UserAuthnCapnpServer struct {
	svc authn.IUserAuthn
}

func (capsrv *UserAuthnCapnpServer) GetProfile(
	ctx context.Context, call hubapi.CapUserAuthn_getProfile) error {

	profile, err := capsrv.svc.GetProfile(ctx)
	if err == nil {
		resp, _ := call.AllocResults()
		profileCapnp := capserializer.MarshalUserProfile(profile)
		err = resp.SetProfile(profileCapnp)
	}
	return err
}

func (capsrv *UserAuthnCapnpServer) Login(
	ctx context.Context, call hubapi.CapUserAuthn_login) error {
	args := call.Args()
	password, _ := args.Password()
	accessToken, refreshToken, err := capsrv.svc.Login(ctx, password)
	if err == nil {
		res, _ := call.AllocResults()
		err = res.SetAuthToken(accessToken)
		_ = res.SetRefreshToken(refreshToken)
	}
	return err
}

func (capsrv *UserAuthnCapnpServer) Logout(
	ctx context.Context, call hubapi.CapUserAuthn_logout) error {
	args := call.Args()
	refreshToken, _ := args.RefreshToken()
	err := capsrv.svc.Logout(ctx, refreshToken)
	return err
}

func (capsrv *UserAuthnCapnpServer) Refresh(
	ctx context.Context, call hubapi.CapUserAuthn_refresh) error {
	args := call.Args()
	refreshToken, _ := args.RefreshToken()
	newAccessToken, newRefreshToken, err := capsrv.svc.Refresh(ctx, refreshToken)
	if err == nil {
		res, _ := call.AllocResults()
		err = res.SetNewAuthToken(newAccessToken)
		_ = res.SetNewRefreshToken(newRefreshToken)
	}
	return err
}

func (capsrv *UserAuthnCapnpServer) SetPassword(
	ctx context.Context, call hubapi.CapUserAuthn_setPassword) error {
	args := call.Args()
	newPassword, _ := args.NewPassword()
	err := capsrv.svc.SetPassword(ctx, newPassword)
	return err
}

func (capsrv *UserAuthnCapnpServer) SetProfile(
	ctx context.Context, call hubapi.CapUserAuthn_setProfile) error {

	args := call.Args()
	profileCapnp, _ := args.NewProfile()
	profile := capserializer.UnmarshalUserProfile(profileCapnp)
	err := capsrv.svc.SetProfile(ctx, profile)
	return err
}
