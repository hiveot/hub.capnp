package authn

import (
	"context"
)

// ServiceName of the service used for logging and connecting
const ServiceName = "authn"

// DefaultAccessTokenValiditySec with access token validity in seconds
const DefaultAccessTokenValiditySec = 3600

// DefaultRefreshTokenValiditySec with Refresh token validity before refresh
const DefaultRefreshTokenValiditySec = 1209600

// UserProfile contains user information
type UserProfile struct {
	// The user's login ID, typically email
	LoginID string
	// The user's presentation name
	Name string
	// Last updated password in unix time
	Updated int64
}

// IAuthn defines the interface for simple user management and authentication
type IAuthn interface {
	// CapUserAuthn provides the authentication capabilities for unauthenticated users.
	CapUserAuthn(ctx context.Context, clientID string) IUserAuthn

	// CapManageAuthn provides the capability manage users for use by administrators.
	CapManageAuthn(ctx context.Context) IManageAuthn
}

// IManageAuthn defines the interface for managing the authentication service
// Intended for administrators only.
type IManageAuthn interface {

	// AddUser adds a new user and generates a temporary password
	AddUser(ctx context.Context, loginID string, name string) (password string, err error)

	// ListUsers provide a list of users and their info
	ListUsers(ctx context.Context) (profiles []UserProfile, err error)

	// RemoveUser removes a user and disables login
	// Existing tokens are immediately expired (tbd)
	RemoveUser(ctx context.Context, loginID string) error

	// ResetPassword reset the user's password and returns a new password
	ResetPassword(ctx context.Context, loginID string) (newPassword string, err error)

	// Release the provided capability after use
	Release()
}

// IUserAuthn defines the capabilities to handle user authentication
// Intended for end-users to login, logout, or obtain their profile
type IUserAuthn interface {

	// GetProfile returns the user's profile
	// Login or Refresh must be called successfully first.
	GetProfile(ctx context.Context) (profile UserProfile, err error)

	// Login to authenticate a user
	// This returns a short lived auth token for use with the HTTP api,
	// and a medium lived refresh token used to obtain a new auth token.
	Login(ctx context.Context, password string) (authToken, refreshToken string, err error)

	// Logout invalidates the refresh token
	Logout(ctx context.Context, refreshToken string) (err error)

	// Refresh an authentication token
	// Refresh can be used instead of Login to authenticate and access the profile
	// refreshToken must be a valid refresh token obtained at login
	// This returns a short lived auth token and medium lived refresh token
	Refresh(ctx context.Context, refreshToken string) (newAuthToken, newRefreshToken string, err error)

	// SetPassword changes the client password
	// Login or Refresh must be called successfully first.
	SetPassword(ctx context.Context, newPassword string) error

	// SetProfile updates the user profile
	// Login or Refresh must be called successfully first.
	SetProfile(ctx context.Context, profile UserProfile) error

	// TBD add OAuth2 login support

	// Release the provided capability after use
	Release()
}
