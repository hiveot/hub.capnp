// Package state with POGS capability to read and write application state
// Unfortunately capnp does generate POGS types so we need to duplicate them
package state

import "context"

// IState defines a POGS based capability API of the state store
type IState interface {

	// CapClientState provides the capability to store state for a client application
	// The caller must verify that the clientID is properly authenticated to ensure the capability
	// is handed out to a valid user.
	CapClientState(ctx context.Context, clientID string, appID string) IClientState
}

// IClientState defines a POGS based capability for reading and writing state values
type IClientState interface {
	// Get returns the document for the given key
	// Returns an error if the key doesn't exist
	Get(ctx context.Context, key string) (value string, err error)

	// Set sets a document with the given key
	Set(ctx context.Context, key string, value string) error
}
