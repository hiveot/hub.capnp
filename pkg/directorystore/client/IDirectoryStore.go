// Package client with POGS definitions of the directory store.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package client

import "context"

// IDirectoryStore defines a POGS based capability API of the thing directory store
type IDirectoryStore interface {

	// GetTD returns the TD document for the given Thing ID in JSON format
	GetTD(ctx context.Context, thingID string) (tdJson string, err error)

	// ListTDs returns all TD documents in JSON format
	ListTDs(ctx context.Context, limit int, offset int) (tds []string, err error)

	// QueryTDs returns the TD's filtered using JSONpath on the TD content
	// See 'docs/query-tds.md' for examples
	QueryTDs(ctx context.Context, jsonPath string, limit int, offset int) (tds []string, err error)

	// RemoveTD removes a TD document from the store
	RemoveTD(ctx context.Context, thingID string) (err error)

	// UpdateTD updates the TD document in the directory
	// If the TD with the given ID doesn't exist it will be added.
	UpdateTD(ctx context.Context, thingID string, tdDoc string) (err error)
}
