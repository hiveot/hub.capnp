package auth

// IAclStoreReader is an ACL reader for client roles and groups for access control
// This interface defines the API for the store.
type IAclStoreReader interface {
	// GetGroups returns a list of groups a thing or user is a member of
	GetGroups(clientID string) []string

	// Get highest role of a user has in a list of group
	// Intended to get client permissions in case of overlapping groups
	// Returns the role
	GetRole(clientID string, groupIDs []string) string

	// Close the store
	Close()

	// Open the store
	Open() error
}

// IAclStoreWriter is a writer to a ACL store for client roles and groups
type IAclStoreWriter interface {
	IAclStoreReader

	// Write the role for the client in a group
	SetRole(clientID string, groupID string, role string)
}
