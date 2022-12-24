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
	_ = clientID
	return dummy
}

func (dummy *DummyAuthnService) CapManageAuthn(_ context.Context) authn.IManageAuthn {
	return dummy
}

// --- Manage ---

func (dummy *DummyAuthnService) AddUser(_ context.Context, loginID string) (password string, err error) {
	dummy.pwMap[loginID] = loginID
	return "newpassword", nil
}

func (dummy *DummyAuthnService) ListUsers(_ context.Context) (profiles []authn.UserProfile, err error) {
	profiles = make([]authn.UserProfile, 0)
	return profiles, nil
}

func (dummy *DummyAuthnService) RemoveUser(_ context.Context, loginID string) error {
	_ = loginID
	return nil
}

func (dummy *DummyAuthnService) ResetPassword(_ context.Context, loginID string) (newPassword string, err error) {
	_ = loginID
	newpw := "newpassword"
	return newpw, nil
}

func (dummy *DummyAuthnService) Release() {
}

func (dummy *DummyAuthnService) GetProfile(_ context.Context) (profile authn.UserProfile, err error) {
	profile = authn.UserProfile{}
	return profile, nil
}

func (dummy *DummyAuthnService) Login(_ context.Context, password string) (authToken, refreshToken string, err error) {
	_ = password
	authToken = "auth"
	refreshToken = "refresh"
	err = nil
	return
}

func (dummy *DummyAuthnService) Logout(_ context.Context, refreshToken string) (err error) {
	_ = refreshToken
	return nil
}

func (dummy *DummyAuthnService) Refresh(_ context.Context, refreshToken string) (newAuthToken, newRefreshToken string, err error) {
	_ = refreshToken
	newAuthToken = "auth"
	newRefreshToken = "refresh"
	err = nil
	return
}

// SetPassword changes the client password
// Login or Refresh must be called successfully first.
func (dummy *DummyAuthnService) SetPassword(_ context.Context, newPassword string) error {
	_ = newPassword
	return nil
}

// SetProfile updates the user profile
// Login or Refresh must be called successfully first.
func (dummy *DummyAuthnService) SetProfile(_ context.Context, profile authn.UserProfile) error {
	_ = profile
	return nil
}

func NewDummyAuthnService() *DummyAuthnService {
	dummy := &DummyAuthnService{
		pwMap: map[string]string{},
	}
	return dummy
}
