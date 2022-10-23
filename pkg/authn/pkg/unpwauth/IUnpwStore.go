package unpwauth

// IUnpwStore defined the interface for accessing the username-password store
type IUnpwStore interface {
	// Close the store
	Close()

	// GetPasswordHash returns the password hash for the user, or "" if the user is not found
	GetPasswordHash(username string) string

	// Open the store
	Open() error

	// SetPasswordHash writes and updates the password for the given user
	//  loginID is the login ID of the user whose hash to write
	//  hash is the calculated password hash to store. This is independent of the hashing algorithm.
	// Returns error if the store isn't writable
	SetPasswordHash(loginID string, hash string) error
}
