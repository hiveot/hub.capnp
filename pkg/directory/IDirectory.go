// Package directory with POGS capability definitions of the directory store.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package directory

import (
	"context"

	"github.com/hiveot/hub/internal/bucketstore"
)

// ServiceName is the name of the service to connect to
const ServiceName = "directory"
const TDBucketName = "td"

// IDirectory defines the capability to use the thing directory
type IDirectory interface {
	// CapReadDirectory provides the capability to read and query the thing directory
	CapReadDirectory(ctx context.Context) IReadDirectory

	// CapUpdateDirectory provides the capability to update the thing directory
	CapUpdateDirectory(ctx context.Context) IUpdateDirectory

	// Stop the service and free its resources
	Stop(ctx context.Context) error
}

// IReadDirectory defines the capability of reading the Thing directory
type IReadDirectory interface {
	// Cursor returns an iterator for TD documents
	Cursor(ctx context.Context) (cursor IDirectoryCursor)

	// GetTD returns the TD document for the given Thing ID in JSON format
	GetTD(ctx context.Context, thingID string) (tdJson string, err error)

	// QueryTDs returns the TD's filtered using JSONpath on the TD content
	// See 'docs/query-tds.md' for examples
	// disabled as this is not used
	//QueryTDs(ctx context.Context, jsonPath string, limit int, offset int) (tds []string, err error)

	// Release this capability and allocated resources after its use
	Release()
}

// IUpdateDirectory defines the capability of updating the Thing directory
type IUpdateDirectory interface {

	// RemoveTD removes a TD document from the store
	RemoveTD(ctx context.Context, thingID string) (err error)

	// UpdateTD updates the TD document in the directory
	// If the TD with the given ID doesn't exist it will be added.
	UpdateTD(ctx context.Context, thingID string, tdDoc string) (err error)

	// Release this capability and allocated resources after its use
	Release()
}

// IDirectoryCursor defines the directory iterator
type IDirectoryCursor interface {
	bucketstore.IBucketCursor
}
