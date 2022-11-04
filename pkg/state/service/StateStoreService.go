package statekvstore

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/bucketstore"
	"github.com/hiveot/hub/internal/bucketstore/bolts"
	"github.com/hiveot/hub/pkg/state"
	"github.com/hiveot/hub/pkg/state/config"
)

// StateStoreService implements the server for storing application state for clients
// This implements the IStateStore interface
// Each client will have its own store instance that is used by all instance of the client store capability.
type StateStoreService struct {
	cfg config.StateConfig
	// The underlying persistence store for each client that is concurrent safe.
	// Released stores have a nil entry
	clientStores map[string]bucketstore.IBucketStore
	// reference count of issued capabilities by serviceID
	// released stores have a 0 value
	clientRefs map[string]int
	//
	mux sync.Mutex
}

// CapClientState returns a new instance of the capability to store client state.
// This uses one store instance per clientID.
func (srv *StateStoreService) CapClientState(
	ctx context.Context, clientID string, bucketID string) (cap state.IClientState, err error) {

	srv.mux.Lock()
	defer srv.mux.Unlock()
	clientStore := srv.clientStores[clientID]
	// create the store instance for the client if one doesn't yet exist
	if clientStore == nil {
		// TODO: backend depend on config
		if srv.cfg.Backend == config.StateBackendKVStore {
			logrus.Infof("opening kv store for client '%s' bucket '%s", clientID, bucketID)
			storePath := path.Join(srv.cfg.StoreDirectory, clientID+".json")
			clientStore = kvmem.NewKVStore(clientID, storePath)
		} else {
			logrus.Infof("opening boltDB store for client '%s' bucket '%s", clientID, bucketID)
			storePath := path.Join(srv.cfg.StoreDirectory, clientID+".boltdb")
			clientStore = bolts.NewBoltBucketStore(clientID, storePath)
		}
		err = clientStore.Open()
		if err == nil {
			srv.clientStores[clientID] = clientStore
		}
	}
	// multiple client instances use the same store
	refCount := srv.clientRefs[clientID]
	refCount++
	srv.clientRefs[clientID] = refCount
	capability := NewClientState(clientStore, clientID, bucketID, srv.onClientReleased)
	return capability, err
}

// callback to remove the client store when all its clients are removed
func (srv *StateStoreService) onClientReleased(clientID string) {
	srv.mux.Lock()
	defer srv.mux.Unlock()
	refCount := srv.clientRefs[clientID]
	if refCount <= 0 {
		logrus.Errorf("Client '%s' released but its refcount is 0", clientID)
	} else {
		refCount--
		srv.clientRefs[clientID] = refCount
	}
	// remove the store if refCount reaches 0
	if refCount == 0 {
		clientStore := srv.clientStores[clientID]
		if clientStore == nil {
			logrus.Errorf("Client '%s' released but it doesn't have a store", clientID)
		} else {
			srv.clientStores[clientID] = nil
			clientStore.Close()
		}
	}
}

// Start the store
// Ensure the store location exists and is writable
func (srv *StateStoreService) Start() error {
	logrus.Infof("starting state store service")
	info, err := os.Stat(srv.cfg.StoreDirectory)
	if err != nil {
		if !os.IsNotExist(err) {
			// not sure whats wrong here but the service can't continue
			logrus.Errorf("store directory exists but can't be used", err)
			return err
		}
		err = os.MkdirAll(srv.cfg.StoreDirectory, 0700)
		if err != nil {
			// unable to create the store directory
			logrus.Errorf("unable to create the store directory:%s", err)
			return err
		}
	} else if !info.IsDir() {
		err = fmt.Errorf("'%s' is not a directory", srv.cfg.StoreDirectory)
		return err
	}

	return err
}

// Stop releases each of the client capabilities
func (srv *StateStoreService) Stop() error {
	logrus.Infof("stopping state store service")
	// build the list of stores to close to allow list update while closing
	srv.mux.Lock()
	clientList := make([]string, 0, len(srv.clientStores))
	for clientID, store := range srv.clientStores {
		if store != nil {
			clientList = append(clientList, clientID)
		}
	}
	srv.mux.Unlock()

	// Note that closing the store without releasing its clients can result in calls after
	// its store is closed. The store should be able to handle this.
	for _, clientID := range clientList {
		srv.mux.Lock()
		clientStore := srv.clientStores[clientID]
		srv.clientStores[clientID] = nil
		srv.mux.Unlock()
		if clientStore != nil {
			logrus.Infof("Stopping store for '%s'", clientID)
			clientStore.Close()
		}
	}
	return nil
}

// NewStateStoreService creates a state storage server instance.
// Intended for use by the launcher.
//  storeDirectory is the location to store client state files
func NewStateStoreService(stateConfig config.StateConfig) *StateStoreService {
	// TODO: use configuration to determine backend and load limits
	//kvStore := kvstore.NewKVStore(stateConfig.DatabaseURL)
	srv := &StateStoreService{
		cfg:          stateConfig,
		clientStores: make(map[string]bucketstore.IBucketStore),
		clientRefs:   make(map[string]int),
		mux:          sync.Mutex{},
	}
	return srv
}
