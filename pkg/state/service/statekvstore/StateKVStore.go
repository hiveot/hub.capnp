package statekvstore

import (
	"context"

	"github.com/hiveot/hub/internal/kvstore"
)

// StateKVStoreServer is a wrapper around the internal KVStore
// This implements the IState interface
type StateKVStoreServer struct {
	store *kvstore.KVStore
}

// Get returns the document for the given key
// The document can be any text.
func (srv *StateKVStoreServer) Get(_ context.Context, key string) (value string, err error) {
	val, err := srv.store.Read(key)
	return val, err
}

// Set writes a document with the given key
func (srv *StateKVStoreServer) Set(ctx context.Context, key string, value string) error {
	err := srv.store.Write(key, value)
	return err
}

// Stop the storage server and flush changes to disk
func (srv *StateKVStoreServer) Stop() {
	srv.store.Stop()
}

// NewStateKVStoreServer creates a state storage server instance
//  stateStorePath is the file holding the state store data
func NewStateKVStoreServer(stateStorePath string) (*StateKVStoreServer, error) {
	// TODO: use configuration to determine backend and load limits
	kvStore, err := kvstore.NewKVStore(stateStorePath)
	srv := &StateKVStoreServer{
		store: kvStore,
	}
	err = kvStore.Start()
	return srv, err
}
