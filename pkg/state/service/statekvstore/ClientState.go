package statekvstore

import (
	"context"
	"fmt"

	"github.com/hiveot/hub/internal/kvstore"
)

// ClientState is a store capability for a specific client application
type ClientState struct {
	// The underlying persistence store that is concurrent safe
	store *kvstore.KVStore
	// The client whose state to store
	clientID string
	// The application ID whose state to store
	appID string
}

// Close capability and release resources, if any
func (svc *ClientState) Release() {
}

// Get returns the document for the given key
// The document can be any text.
func (srv *ClientState) Get(_ context.Context, key string) (value string, err error) {
	clientKey := fmt.Sprintf("%s.%s.%s", srv.clientID, srv.appID, key)
	val, err := srv.store.Read(clientKey)
	return val, err
}

// Set writes a document with the given key
func (srv *ClientState) Set(_ context.Context, key string, value string) error {
	clientKey := fmt.Sprintf("%s.%s.%s", srv.clientID, srv.appID, key)
	err := srv.store.Write(clientKey, value)
	return err
}

// NewClientStateKVStore creates a new instance for storing a client's application state
func NewClientStateKVStore(store *kvstore.KVStore, clientID string, appID string) *ClientState {
	cl := &ClientState{
		store:    store,
		clientID: clientID,
		appID:    appID,
	}
	return cl
}
