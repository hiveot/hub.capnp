package statekvstore

import (
	"context"
	"fmt"

	"github.com/hiveot/hub/internal/kvstore"
)

// ClientStateKVStore is a store instance for a specific client application
type ClientStateKVStore struct {
	// The underlying persistence store that is concurrent safe
	store *kvstore.KVStore
	// The client whose state to store
	clientID string
	// The application ID whose state to store
	appID string
}

// Get returns the document for the given key
// The document can be any text.
func (srv *ClientStateKVStore) Get(_ context.Context, key string) (value string, err error) {
	clientKey := fmt.Sprintf("%s.%s.%s", srv.clientID, srv.appID, key)
	val, err := srv.store.Read(clientKey)
	return val, err
}

// Set writes a document with the given key
func (srv *ClientStateKVStore) Set(_ context.Context, key string, value string) error {
	clientKey := fmt.Sprintf("%s.%s.%s", srv.clientID, srv.appID, key)
	err := srv.store.Write(clientKey, value)
	return err
}

// NewClientStateKVStore creates a new instance for storing a client's application state
func NewClientStateKVStore(store *kvstore.KVStore, clientID string, appID string) *ClientStateKVStore {
	cl := &ClientStateKVStore{
		store:    store,
		clientID: clientID,
		appID:    appID,
	}
	return cl
}
