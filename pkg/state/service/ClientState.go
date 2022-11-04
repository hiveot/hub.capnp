package statekvstore

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/bucketstore"
	"github.com/hiveot/hub/pkg/state"
)

// ClientState is a store capability for a specific client bucket
// this is just a wrapper with a release callback. To be removed
// Multiple client instances can use the same bucket store
type ClientState struct {
	// The underlying persistence store that is concurrent safe
	store bucketstore.IBucketStore
	// The client whose state to store
	clientID string
	// The bucket to store state into. Can be application ID or other
	bucketID string
	// callback to invoke when release is called
	onReleaseCB func(clientID string)
}

// Delete a key from the store
func (svc *ClientState) Delete(ctx context.Context, key string) (err error) {
	err = svc.store.Delete(svc.bucketID, key)
	return err
}

// Get returns the document for the given key
// The document can be any text.
func (svc *ClientState) Get(ctx context.Context, key string) (value string, err error) {
	val, err := svc.store.Get(svc.bucketID, key)
	return val, err
}

// GetMultiple returns a batch of documents for the given key
// The document can be any text.
func (svc *ClientState) GetMultiple(
	ctx context.Context, keys []string) (docs map[string]string, err error) {

	docs, err = svc.store.GetMultiple(svc.bucketID, keys)
	return docs, err
}

// Release capability and its resources, if any
func (svc *ClientState) Release() {
	// callback can be used for reference counting to the bucket store.
	logrus.Infof("Client '%s' released", svc.clientID)
	svc.onReleaseCB(svc.clientID)
}

// Set writes a document with the given key
func (svc *ClientState) Set(ctx context.Context, key string, value string) error {
	err := svc.store.Set(svc.bucketID, key, value)
	return err
}

// SetMultiple writes a batch of key-values
func (svc *ClientState) SetMultiple(ctx context.Context, docs map[string]string) (err error) {
	err = svc.store.SetMultiple(svc.bucketID, docs)
	return err
}

// NewClientState creates a new instance for storing a client's application state
//  onRelease callback to invoke when the client is released by its protocol binding
func NewClientState(store bucketstore.IBucketStore,
	clientID string, bucketID string,
	onReleaseCB func(clientID string)) state.IClientState {
	cl := &ClientState{
		store:       store,
		clientID:    clientID,
		bucketID:    bucketID,
		onReleaseCB: onReleaseCB,
	}
	return cl
}
