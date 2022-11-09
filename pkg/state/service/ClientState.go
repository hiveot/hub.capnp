package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/bucketstore"
	"github.com/hiveot/hub/pkg/state"
)

// ClientState is a store capability to read and write a specific client bucket
// this is just a wrapper with a release callback. To be removed
// Multiple client instances can use the same bucket store
type ClientState struct {
	// The underlying persistence bucket that is concurrent safe
	bucket bucketstore.IBucket
	// The client whose state to store
	clientID string
	// The bucket to store state into. Can be application ID or other
	bucketID string
	// callback to invoke when release is called so the owning store can clean up
	onReleaseCB func(clientID string)
}

// Cursor provides an iterator cursor for the bucket
func (svc *ClientState) Cursor(ctx context.Context) (cursor state.IClientCursor, err error) {
	cursor, err = svc.bucket.Cursor()
	return cursor, err
}

// Delete a key from the bucket
func (svc *ClientState) Delete(ctx context.Context, key string) (err error) {
	err = svc.bucket.Delete(key)
	return err
}

// Get returns the document for the given key
// The document can be any text.
func (svc *ClientState) Get(ctx context.Context, key string) (value []byte, err error) {
	val, err := svc.bucket.Get(key)
	return val, err
}

// GetMultiple returns a batch of documents for the given key
// The document can be any text.
func (svc *ClientState) GetMultiple(
	ctx context.Context, keys []string) (docs map[string][]byte, err error) {
	docs = make(map[string][]byte)
	docs, err = svc.bucket.GetMultiple(keys)
	return docs, err
}

// Release capability and the bucket
// This invokes the callback after closing the bucket
func (svc *ClientState) Release() {
	// callback can be used for reference counting to the bucket store.
	logrus.Infof("Releasing client '%s' bucket '%s'", svc.clientID, svc.bucketID)
	// client state buckets are always writable
	err := svc.bucket.Close()
	if err != nil {
		logrus.Errorf("closing bucket did not go as planned: %s", err)
	}
	svc.onReleaseCB(svc.clientID)
}

// Set writes a document with the given key
func (svc *ClientState) Set(ctx context.Context, key string, value []byte) error {
	err := svc.bucket.Set(key, value)
	return err
}

// SetMultiple writes a batch of key-values
func (svc *ClientState) SetMultiple(ctx context.Context, docs map[string][]byte) (err error) {
	err = svc.bucket.SetMultiple(docs)
	return err
}

// NewClientState creates a new instance for storing a client's application state
//  clientID
//  bucketID
//  bucket to store data in
//  onRelease callback to invoke when the client is released by its protocol binding
func NewClientState(
	clientID string, bucketID string,
	bucket bucketstore.IBucket,
	onReleaseCB func(clientID string)) state.IClientState {
	cl := &ClientState{
		bucket:      bucket,
		clientID:    clientID,
		bucketID:    bucketID,
		onReleaseCB: onReleaseCB,
	}
	return cl
}
