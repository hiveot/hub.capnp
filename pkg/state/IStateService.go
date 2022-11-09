// Package state with capability to persist application state in a KV store.
package state

import (
	"context"

	"github.com/hiveot/hub/internal/bucketstore"
)

const ServiceName = "state"

// CollectionStatus holds the status of the requested collection
//type CollectionStatus struct {
//	Name            string // name of the collection,
//	ErrorCount      int    // total commit error count for the collection, if available
//	MaxSize         int64  // Maximum allowed memory size in bytes to store
//	MaxDocumentSize int    // Maximum allowed size of a single document in bytes
//	MaxDocuments    int    // Maximume allowed number of documents in the collection
//	NrDocuments     int    // Current number of documents in the collection, eg nr clients, nr applications, nr documents
//	Size            int    // Estimated memory size in bytes is by the collection
//}

// IStateService defines the capability API of the general purpose state store for service storage needs.
// This store is backed by the embedded BBolt KV store.
//
// While the store limitations depend on the platform resources, the following limitations should work
// with most platforms, eg raspberry pi 4.
// * Total DB size: 4GB, or more depending on available disk space (tbd)
// * Number of clients: limited by disk space
// * Number of applications (tables/buckets) per client: 1000 (arbitrary )
// * Document size: less than 100KB per document (tbd)
// * Number of documents per client: 10000 (tbd)
//
type IStateService interface {

	// GetStores the names of available stores for use in application state
	//GetStores(ctx context.Context) []string

	// CapBucketStore provides the capability to access storage for a client
	//CapBucketStore(ctx context.Context, clientID string) bucketstore.IBucketStore

	// CapClientState provides the capability to store and retrieve application state.
	// The caller must verify that the clientID is properly authenticated to ensure the
	// capability is handed out to a valid client.
	// Buckets allow clients to have multiple stores with different types of data.
	//
	//  clientID is the service that represents the application. Each client uses a separate DB.
	//  bucketID is the name of the bucket in the store. Eg a separate table.
	CapClientBucket(ctx context.Context, clientID string, bucketID string) (cap IClientState, err error)

	// CapManageStateStore provides the capability to manage the state store
	//CapManageStateStore(ctx context.Context) IManageStateStore

	// Stop the service and free its resources
	Stop() error
}

// IManageStateStore defines a capability to manage the state store
type IManageStateStore interface {
	// Status returns the overall state store info
	//GetStatus(ctx context.Context) CollectionStatus

	// GetClientStatus returns the info of a specific client
	//GetClientStatus(ctx context.Context, clientID string) CollectionStatus

	// Release the ManageStateStore capability and release its resources
	Release()
}

// IClientState defines the capability for reading and writing state values in a storage bucket
type IClientState interface {

	// Cursor creates a new cursor for iterating the content of the client bucket
	// cursor.Close must be called after use to release any read transactions
	// returns an error the cursor cannot be created
	Cursor(ctx context.Context) (cap IClientCursor, err error)

	// Delete removes the key-value pair from the state store
	Delete(ctx context.Context, key string) (err error)

	// Get returns the document for the given key
	// Returns an error if the key doesn't exist
	Get(ctx context.Context, key string) (value []byte, err error)

	// GetMultiple returns a batch of documents with the given keys
	GetMultiple(ctx context.Context, keys []string) (docs map[string][]byte, err error)

	// Keys returns a list of document keys in the store
	//Keys(ctx context.Context) (keys []string, err error)

	// Set sets a document with the given key
	Set(ctx context.Context, key string, value []byte) error

	// SetMultiple sets multiple documents in a batch update
	SetMultiple(ctx context.Context, docs map[string][]byte) (err error)

	// Status returns the application state status
	//Status(ctx context.Context) CollectionStatus

	// Release the ClientState capability and free its resources
	Release()
}

type IClientCursor interface {
	bucketstore.IBucketCursor
}
