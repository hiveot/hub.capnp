package dummy

import (
	"context"

	"github.com/hiveot/hub/pkg/authn"
)

// DummyAuthnService for testing. This implements IAuthnService, IManageAuthn and IUserAuthn
type DummyAuthnService struct {
	pwMap map[string]string
}

func (dummy *DummyAuthnService) CapUserAuthn(_ context.Context, clientID string) authn.IUserAuthn {
	return dummy
}

func (dummy *DummyAuthnService) CapManageAuthn(_ context.Context) authn.IManageAuthn {
	return dummy
}

// --- Manage ---

func (dummy *DummyAuthnService) AddUser(_ context.Context, loginID string, name string) (password string, err error) {
	dummy.pwMap[loginID] = name
	return "newpassword", nil
}

func (dummy *DummyAuthnService) ListUsers(_ context.Context) (profiles []authn.UserProfile, err error) {
	profiles = make([]authn.UserProfile, 0)
	return profiles, nil
}

func (dummy *DummyAuthnService) RemoveUser(_ context.Context, loginID string) error {
	return nil
}

func (dummy *DummyAuthnService) ResetPassword(_ context.Context, loginID string) (newPassword string, err error) {
	newpw := "newpassword"
	return newpw, nil
}

func (dummy *DummyAuthnService) Release() {
}

func (dummy *DummyAuthnService) GetProfile(ctx context.Context) (profile authn.UserProfile, err error) {
	profile = authn.UserProfile{}
	return profile, nil
}

func (dummy *DummyAuthnService) Login(ctx context.Context, password string) (authToken, refreshToken string, err error) {
	authToken = "auth"
	refreshToken = "refresh"
	err = nil
	return
}

func (dummy *DummyAuthnService) Logout(ctx context.Context, refreshToken string) (err error) {
	return nil
}

func (dummy *DummyAuthnService) Refresh(ctx context.Context, refreshToken string) (newAuthToken, newRefreshToken string, err error) {
	newAuthToken = "auth"
	newRefreshToken = "refresh"
	err = nil
	return
}

// SetPassword changes the client password
// Login or Refresh must be called successfully first.
func (dummy *DummyAuthnService) SetPassword(ctx context.Context, newPassword string) error {
	return nil
}

// SetProfile updates the user profile
// Login or Refresh must be called successfully first.
func (dummy *DummyAuthnService) SetProfile(ctx context.Context, profile authn.UserProfile) error {
	return nil
}

func NewDummyAuthnService() *DummyAuthnService {
	dummy := &DummyAuthnService{
		pwMap: map[string]string{},
	}
	return dummy
}
