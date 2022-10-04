package statekvstore

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/kvstore"
	"github.com/hiveot/hub/pkg/state"
)

// StateKVStore implements the server for storing application state
// This implements the IState interface
type StateKVStore struct {
	// The underlying persistence store that is concurrent safe
	store *kvstore.KVStore
}

// CapClientState returns the capability to store client application state
func (srv *StateKVStore) CapClientState(_ context.Context, clientID string, appID string) state.IClientState {
	capability := NewClientStateKVStore(srv.store, clientID, appID)
	return capability
}

// Stop the store and flush changes to disk
func (srv *StateKVStore) Stop() {
	logrus.Infof("stopping state store")
	srv.store.Stop()
}

// NewStateKVStore creates a state storage server instance.
// Intended for use by the launcher.
//  stateStorePath is the file holding the state store data
func NewStateKVStore(stateStorePath string) (*StateKVStore, error) {
	// TODO: use configuration to determine backend and load limits
	kvStore, err := kvstore.NewKVStore(stateStorePath)
	srv := &StateKVStore{
		store: kvStore,
	}
	err = kvStore.Start()
	return srv, err
}
