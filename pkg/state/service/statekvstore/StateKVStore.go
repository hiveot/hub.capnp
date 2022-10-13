package statekvstore

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/kvstore"
	"github.com/hiveot/hub/pkg/state"
	"github.com/hiveot/hub/pkg/state/config"
)

// StateKVStore implements the server for storing application state
// This implements the IState interface
type StateKVStore struct {
	cfg config.StateConfig
	// The underlying persistence store that is concurrent safe
	store *kvstore.KVStore
}

// CapClientState returns the capability to store client application state
func (srv *StateKVStore) CapClientState(_ context.Context, clientID string, appID string) state.IClientState {
	capability := NewClientStateKVStore(srv.store, clientID, appID)
	return capability
}

// Stop the store and flush changes to disk
func (srv *StateKVStore) Stop() error {
	logrus.Infof("stopping state store")
	err := srv.store.Stop()
	return err
}

// NewStateKVStore creates a state storage server instance.
// Intended for use by the launcher.
//  stateStorePath is the file holding the state store data
func NewStateKVStore(stateConfig config.StateConfig) (*StateKVStore, error) {
	// TODO: use configuration to determine backend and load limits
	kvStore, err := kvstore.NewKVStore(stateConfig.DatabaseURL)
	srv := &StateKVStore{
		cfg:   stateConfig,
		store: kvStore,
	}
	err = kvStore.Start()
	return srv, err
}
