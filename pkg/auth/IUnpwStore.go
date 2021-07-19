package auth

// IUnpwStoreReader defined the interface for reading the username-password store for user authentication
type IUnpwStoreReader interface {
	// Return the password hash for the user, or "" if the user is not found
	GetPasswordHash(username string) string

	// Close the store
	Close()

	// Open the store
	Open() error
}

// IUnpwStoreWriter defines the writer interface for a username-password store for user authentication
type IUnpwStoreWriter interface {
	IUnpwStoreReader

	// Write and update the password for the given user
	//  loginID is the login ID of the user whose hash to write
	//  hash is the calculated password hash to store. This is independent of the hashing algorithm.
	// Returns error if the store isn't writable
	SetPasswordHash(loginID string, hash string) error
}
