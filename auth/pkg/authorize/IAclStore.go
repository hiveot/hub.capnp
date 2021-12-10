package authorize

// IAclStore defines the interface of a group based ACL store
type IAclStore interface {
	// GetGroups returns a list of groups a thing or user is a member of
	GetGroups(clientID string) []string

	// GetRole returns the highest role of a user has in a list of group
	// Intended to get client permissions in case of overlapping groups
	// Returns the role
	GetRole(clientID string, groupIDs []string) string

	// Close the store
	Close()

	// Open the store
	Open() error

	// SetRole writes the role for the client in a group
	SetRole(clientID string, groupID string, role string) error
}
