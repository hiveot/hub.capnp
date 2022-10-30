package service

import (
	"context"
	"fmt"

	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/service/jwtauthn"
	"github.com/hiveot/hub/pkg/authn/service/unpwstore"
)

// UserAuthn provides the client authentication capability for unauthented users
// This implements the IUserAuthn interface
type UserAuthn struct {
	// JWT token handling
	jwtAuthn     *jwtauthn.JWTAuthn
	pwStore      unpwstore.IUnpwStore
	loginID      string // restricted scope of this service
	refreshToken string // set when authenticated
}

// GetProfile returns the user's profile
// User must be authenticated first
func (svc *UserAuthn) GetProfile(ctx context.Context) (profile authn.UserProfile, err error) {
	if svc.refreshToken == "" {
		return profile, fmt.Errorf("not authenticated")
	}
	//upa.profileStore[profile.LoginID] = profile
	entry, err := svc.pwStore.GetEntry(svc.loginID)
	if err == nil {
		profile.LoginID = entry.LoginID
		profile.Name = entry.UserName
		profile.Updated = entry.Updated
	}
	return profile, err

}

// Login to authenticate a user
// This returns a short lived auth token for use with the HTTP api,
// and a medium lived refresh token used to obtain a new auth token.
func (svc *UserAuthn) Login(ctx context.Context, password string) (
	authToken, refreshToken string, err error) {
	_ = ctx
	err = svc.pwStore.VerifyPassword(svc.loginID, password)
	if err != nil {
		return "", "", fmt.Errorf("invalid login as '%s'", svc.loginID)
	}
	// when valid, provide the tokens
	at, rt, err := svc.jwtAuthn.CreateTokens(svc.loginID)
	if err == nil {
		svc.refreshToken = rt
	}
	return at, rt, err
}

// Logout invalidates the refresh token
func (svc *UserAuthn) Logout(ctx context.Context, refreshToken string) (err error) {
	_ = ctx
	// logout is idempotent, same result on repeated calls
	if svc.refreshToken == "" {
		return nil
	}
	// when valid, invalidate the tokens
	svc.refreshToken = ""
	svc.jwtAuthn.InvalidateToken(svc.loginID, refreshToken)
	return nil
}

// Refresh an authentication token
// refreshToken must be a valid refresh token obtained at login
// This returns a short lived auth token and medium lived refresh token
func (svc *UserAuthn) Refresh(ctx context.Context, refreshToken string) (
	newAuthToken, newRefreshToken string, err error) {
	_ = ctx

	at, rt, err := svc.jwtAuthn.RefreshTokens(svc.loginID, refreshToken)
	if err == nil {
		svc.refreshToken = rt
	}
	return at, rt, err
}

// Release the instance and release resource
func (svc *UserAuthn) Release() {
	//svc.loginID = ""
	svc.refreshToken = ""
}

// SetPassword changes the client password
func (svc *UserAuthn) SetPassword(ctx context.Context, newPassword string) error {
	if svc.refreshToken == "" {
		return fmt.Errorf("not authenticated")
	}
	return svc.pwStore.SetPassword(svc.loginID, newPassword)
}

// SetProfile replaces the user profile
func (svc *UserAuthn) SetProfile(ctx context.Context, profile authn.UserProfile) error {
	if svc.refreshToken == "" {
		return fmt.Errorf("not authenticated")
	}
	if profile.LoginID != svc.loginID {
		return fmt.Errorf("profile doesn't match the user's login ID")
	}
	return svc.pwStore.SetName(profile.LoginID, profile.Name)

}
func NewUserAuthn(loginID string, jwtAuthn *jwtauthn.JWTAuthn, pwStore unpwstore.IUnpwStore) *UserAuthn {
	clauthn := &UserAuthn{
		loginID:  loginID,
		jwtAuthn: jwtAuthn,
		pwStore:  pwStore,
	}
	return clauthn
}
