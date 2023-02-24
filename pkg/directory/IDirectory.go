// Package directory with POGS capability definitions of the directory store.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package directory

import (
	"context"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/thing"
)

// ServiceName is the name of the service to connect to
const ServiceName = hubapi.DirectoryServiceName
const TDBucketName = "td"

// IDirectory defines the capability to use the thing directory
type IDirectory interface {

	// CapReadDirectory provides the capability to read and query the thing directory
	CapReadDirectory(ctx context.Context, clientID string) (IReadDirectory, error)

	// CapUpdateDirectory provides the capability to update the thing directory
	CapUpdateDirectory(ctx context.Context, clientID string) (IUpdateDirectory, error)
}

// IDirectoryCursor is a cursor to iterate the directory
type IDirectoryCursor interface {
	// First return the first directory entry.
	//  ValueJSON contains the JSON encoded TD document
	// Returns nil if the store is empty
	First() (thingValue *thing.ThingValue, valid bool)

	// Next returns the next directory entry
	// Returns nil when trying to read past the last value
	Next() (thingValue *thing.ThingValue, valid bool)

	// NextN returns a batch of next directory entries
	// Returns empty list when trying to read past the last value
	// itemsRemaining is true as long as more items can be retrieved
	NextN(steps uint) (batch []*thing.ThingValue, itemsRemaining bool)

	// Release the cursor and resources
	Release()
}

// IReadDirectory defines the capability of reading the Thing directory
type IReadDirectory interface {
	// Cursor returns an iterator for ThingValue objects containing TD documents
	Cursor(ctx context.Context) (cursor IDirectoryCursor)

	// GetTD returns the TD document for the given Publisher/Thing ID in JSON format.
	// Returns the thingValue containing the JSON serialized TD,
	// or nil if the publisherID/thingID doesn't exist and an error if the store is not reachable.
	GetTD(ctx context.Context, publisherID, thingID string) (tv *thing.ThingValue, err error)

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
	RemoveTD(ctx context.Context, publisherID, thingID string) (err error)

	// UpdateTD updates the TD document in the directory
	// If the TD doesn't exist it will be added.
	//  tv is a ThingValue object containing the JSON serialized TD document
	UpdateTD(ctx context.Context, publisherID, thingID string, tdDoc []byte) (err error)

	// Release this capability and allocated resources after its use
	Release()
}
