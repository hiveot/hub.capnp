package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/service/unpwstore"
)

// ManageAuthn provides authentication management services for administrators
// This implements the IManageAuthn interface
type ManageAuthn struct {
	pwStore unpwstore.IUnpwStore
}

// AddUser adds a new user and returns a generated password
func (svc *ManageAuthn) AddUser(ctx context.Context, loginID string, newPassword string) (password string, err error) {
	_ = ctx
	exists := svc.pwStore.Exists(loginID)
	if exists {
		return "", fmt.Errorf("user with loginID '%s' already exists", loginID)
	}
	if newPassword == "" {
		newPassword = svc.GeneratePassword(0, false)
	}
	err = svc.pwStore.SetPassword(loginID, newPassword)
	return newPassword, err
}

// GeneratePassword with upper, lower, numbers and special characters
func (svc *ManageAuthn) GeneratePassword(length int, useSpecial bool) (password string) {
	const charsLow = "abcdefghijklmnopqrstuvwxyz"
	const charsUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const charsSpecial = "!#$%&*+-./:=?@^_"
	const numbers = "0123456789"
	var pool = []rune(charsLow + numbers + charsUpper)

	if length < 2 {
		length = 8
	}
	if useSpecial {
		pool = append(pool, []rune(charsSpecial)...)
	}
	rand.Seed(time.Now().Unix())
	//pwchars := make([]string, length)
	pwchars := strings.Builder{}

	for i := 0; i < length; i++ {
		pos := rand.Intn(len(pool))
		pwchars.WriteRune(pool[pos])
	}
	password = pwchars.String()
	return password
}

// ListUsers provide a list of users and their info
func (svc *ManageAuthn) ListUsers(ctx context.Context) (profiles []authn.UserProfile, err error) {
	_ = ctx
	pwEntries, err := svc.pwStore.List()
	profiles = make([]authn.UserProfile, len(pwEntries))
	for i, entry := range pwEntries {
		profile := authn.UserProfile{
			LoginID: entry.LoginID,
			Name:    entry.UserName,
			Updated: entry.Updated,
		}
		profiles[i] = profile
	}
	return profiles, err
}

// Release the capability after use
func (svc *ManageAuthn) Release() {
	// nothing to do here
}

// RemoveUser removes a user and disables login
// Existing tokens are immediately expired (tbd)
func (svc *ManageAuthn) RemoveUser(ctx context.Context, loginID string) (err error) {
	_ = ctx
	err = svc.pwStore.Remove(loginID)
	return err
}

// ResetPassword reset a user's password and returns a new temporary password
func (svc *ManageAuthn) ResetPassword(ctx context.Context, loginID string, newPassword string) (password string, err error) {
	_ = ctx
	if newPassword == "" {
		newPassword = svc.GeneratePassword(8, false)
	}
	err = svc.pwStore.SetPassword(loginID, newPassword)
	return newPassword, err
}

// UpdateUser updates a user's name
func (svc *ManageAuthn) UpdateUser(ctx context.Context, loginID string, name string) (err error) {
	_ = ctx
	exists := svc.pwStore.Exists(loginID)
	if !exists {
		return fmt.Errorf("user with loginID '%s' does not exist", loginID)
	}
	err = svc.pwStore.SetName(loginID, name)
	return err
}
func NewManageAuthn(pwStore unpwstore.IUnpwStore) *ManageAuthn {
	ma := &ManageAuthn{
		pwStore: pwStore,
	}
	return ma
}
