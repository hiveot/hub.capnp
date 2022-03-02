package authenticate

import (
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// supported password hashes
const (
	PWHASH_ARGON2id = "argon2id"
	PWHASH_BCRYPT   = "bcrypt" // fallback in case argon2i cannot be used
)

// VerifyUsernamePassword is an interface to verify username/password authentication
type VerifyUsernamePassword func(userID string, password string) bool

// Authenticator manages client username/password authentication for access to Things
type Authenticator struct {
	unpwStore IUnpwStore
}

// CreatePasswordHash for the given password
// This creates the hash and does not update the store. See also VerifyPasswordHash
// The only two hashes allowed are argon2id and bcrypt, although argon2id is recommended
//  password to hash
//  algo is the algorithm to use, PWHASH_ARGON2id (default) or PWHASH_BCRYPT
//  iterations for argon2id, default is 10
func CreatePasswordHash(password string, algo string, iterations uint) (hash string, err error) {
	if password == "" {
		return "", fmt.Errorf("CreatePasswordHash: Missing password")
	}
	if algo == "" {
		algo = PWHASH_ARGON2id
	}
	if algo == PWHASH_ARGON2id {
		if iterations <= 0 {
			iterations = 10
		}
		params := argon2id.DefaultParams
		params.Iterations = uint32(iterations)
		hash, err = argon2id.CreateHash(password, params)
	} else if algo == PWHASH_BCRYPT {
		var hashBytes []byte
		hashBytes, err = bcrypt.GenerateFromPassword([]byte(password), 0)
		hash = string(hashBytes)
	} else {
		err = fmt.Errorf("CreatePasswordHash: Unsupported hashing algorithm '%s'", algo)
	}
	return hash, err
}

// SetPassword hashes the given password and stores it in the password store
// Returns if username or password are not provided
func (ah *Authenticator) SetPassword(username string, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("SetPassword: Missing username or password")
	}
	// use default hashing algo
	hash, err := CreatePasswordHash(password, "", 0)
	if err != nil {
		return err
	}
	if ah.unpwStore != nil {
		err = ah.unpwStore.SetPasswordHash(username, hash)
	}
	return err
}

// Start the authhandler. This opens the password store.
// if no password store was provided this simply returns nil
func (ah *Authenticator) Start() error {
	if ah.unpwStore == nil {
		return fmt.Errorf("Authenticator.Start: missing password store")
	}
	err := ah.unpwStore.Open()
	if err != nil {
		err2 := fmt.Errorf("Authenticator.Start Failed opening password store: %s", err)
		logrus.Errorf("%s", err2)
		return err2
	}
	logrus.Infof("Authenticator.Start Success")
	return nil
}

// Stop the authn handler and close the password store.
func (ah *Authenticator) Stop() {
	if ah.unpwStore != nil {
		ah.unpwStore.Close()
	}
}

// VerifyUsernamePassword verifies if the given password is valid for login
// Returns true if valid, false if the user is unknown or the password is invalid
func (ah *Authenticator) VerifyUsernamePassword(loginName string, password string) bool {
	if ah.unpwStore == nil {
		return false
	}

	// Todo: configure hashing method
	algo := PWHASH_ARGON2id
	h := ah.unpwStore.GetPasswordHash(loginName)
	match := ah.VerifyPasswordHash(h, password, algo)
	logrus.Infof("VerifyUsernamePassword: loginName=%s, match=%v", loginName, match)
	return match
}

// VerifyPasswordHash verifies if the given hash matches the password
// This does not access the store
//  hash to verify
//  password to verify against
//  algo is the algorithm to use, PWHASH_ARGON2id or PWHASH_BCRYPT
// returns true if the password matches the hash, or false on mismatch
func (ah *Authenticator) VerifyPasswordHash(hash string, password string, algo string) bool {
	if algo == PWHASH_ARGON2id {
		match, _ := argon2id.ComparePasswordAndHash(password, hash)
		return match
	} else if algo == PWHASH_BCRYPT {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		return (err == nil)
	}
	return false
}

// NewAuthenticator creates a new instance of the authentication handler to update and verify user passwords.
//  unpwStore provides the functions to access the password store.
func NewAuthenticator(unpwStore IUnpwStore) *Authenticator {
	a := Authenticator{
		unpwStore: unpwStore,
	}
	return &a
}
