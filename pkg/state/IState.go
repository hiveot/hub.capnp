// Package state with POGS capability to read and write application state
// Unfortunately capnp does generate POGS types so we need to duplicate them
package state

import "context"

// IState defines a POGS based capability API of the state store
type IState interface {

	// Get returns the document for the given key
	// Returns an error if the key doesn't exist
	Get(ctx context.Context, key string) (value string, err error)

	// Set sets a document with the given key
	Set(ctx context.Context, key string, value string) error
}
