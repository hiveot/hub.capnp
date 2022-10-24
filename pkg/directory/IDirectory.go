// Package directory with POGS capability definitions of the directory store.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package directory

import "context"

// ServiceName is the name of the service to connect to
const ServiceName = "directory"

// IDirectory defines a POGS based capability API of the thing directory
type IDirectory interface {
	// CapReadDirectory provides the capability to read and query the thing directory
	CapReadDirectory(ctx context.Context) IReadDirectory

	// CapUpdateDirectory provides the capability to update the thing directory
	CapUpdateDirectory(ctx context.Context) IUpdateDirectory
}

// IReadDirectory defines a POGS based capability of reading the Thing directory
type IReadDirectory interface {

	// GetTD returns the TD document for the given Thing ID in JSON format
	GetTD(ctx context.Context, thingID string) (tdJson string, err error)

	// ListTDs returns all TD documents in JSON format
	ListTDs(ctx context.Context, limit int, offset int) (tds []string, err error)

	// ListTDcb provides batches of TD documents in JSON format via a callback
	// The callback handler will be invoked until isLast is true or the callback returns an error
	ListTDcb(ctx context.Context, cb func(batch []string, isLast bool) error) (err error)

	// QueryTDs returns the TD's filtered using JSONpath on the TD content
	// See 'docs/query-tds.md' for examples
	QueryTDs(ctx context.Context, jsonPath string, limit int, offset int) (tds []string, err error)
}

// IUpdateDirectory defines a POGS based capability of updating the Thing directory
type IUpdateDirectory interface {

	// RemoveTD removes a TD document from the store
	RemoveTD(ctx context.Context, thingID string) (err error)

	// UpdateTD updates the TD document in the directory
	// If the TD with the given ID doesn't exist it will be added.
	UpdateTD(ctx context.Context, thingID string, tdDoc string) (err error)
}
