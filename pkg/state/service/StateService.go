package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/bucketstore/kvbtree"
	"github.com/hiveot/hub/pkg/state"
	"github.com/hiveot/hub/pkg/state/config"
)

// StateService implements the server for storing application state for clients
// This implements the IStateStore interface
// Each client will have its own store instance that is used by all instance of the client store capability.
type StateService struct {
	cfg config.StateConfig
	// The underlying persistence store for each client that is concurrent safe.
	// Released stores have a nil entry
	//  map[clientID]BucketStore
	clientStores map[string]bucketstore.IBucketStore
	// reference count of issued capabilities by serviceID
	// released stores have a 0 value
	clientRefs map[string]int
	//
	running bool
	mux     sync.Mutex
}

// CapClientState returns a new instance of the capability to store client state in a bucket.
// This opens a store for the client if one doesn't yet exist.
func (srv *StateService) CapClientState(
	_ context.Context, clientID string, bucketID string) (state.IClientState, error) {

	logrus.Infof("clientID=%s, bucketID=%s", clientID, bucketID)
	srv.mux.Lock()
	defer srv.mux.Unlock()
	if !srv.running {
		err := fmt.Errorf("state store service has stopped. No new clients allowed.")
		logrus.Error(err)
		return nil, err
	}
	clientStore := srv.clientStores[clientID]
	// create the store instance for the client if one doesn't yet exist
	if clientStore == nil {
		//clientStore = cmd.NewBucketStore(srv.cfg.StoreDirectory, clientID, srv.cfg.Backend)
		//if srv.cfg.Backend == config.StateBackendKVStore {
		logrus.Infof("opening kv store for client '%s' bucket '%s", clientID, bucketID)
		storePath := path.Join(srv.cfg.StoreDirectory, clientID+".json")
		clientStore = kvbtree.NewKVStore(clientID, storePath)
		//} else if srv.cfg.Backend == config.StateBackendBBolt {
		//	logrus.Infof("opening boltDB store for client '%s' bucket '%s", clientID, bucketID)
		//	storePath := path.Join(srv.cfg.StoreDirectory, clientID+".boltdb")
		//	clientStore = bolts.NewBoltStore(clientID, storePath)
		//} else {
		//	logrus.Infof("opening Pebble store for client '%s' bucket '%s", clientID, bucketID)
		//	storePath := path.Join(srv.cfg.StoreDirectory, clientID) // this is a folder
		//	clientStore = pebble.NewPebbleStore(clientID, storePath)
		//}
		err := clientStore.Open()
		if err == nil {
			srv.clientStores[clientID] = clientStore
		}
	}
	// multiple bucket instances use the same store
	// refCount keeps track how many buckets are outstanding for this client.
	refCount := srv.clientRefs[clientID]
	refCount++
	srv.clientRefs[clientID] = refCount
	bucket := clientStore.GetBucket(bucketID)
	capability := NewClientState(clientID, bucketID, bucket, srv.onClientReleased)
	return capability, nil
}

// callback to close the client store when all its clients are removed
// An error is reported if not all buckets are running
func (srv *StateService) onClientReleased(clientID string) {
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
			_ = clientStore.Close()
		}
	}
}

// Start the state service
// Ensure the stores location exists and is writable
func (srv *StateService) Start(_ context.Context) error {
	logrus.Infof("starting state store service")
	info, err := os.Stat(srv.cfg.StoreDirectory)
	if err != nil {
		if !os.IsNotExist(err) {
			// not sure whats wrong here but the service can't continue
			logrus.Errorf("store directory exists but can't be used: %s", err)
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
	} else if err = unix.Access(srv.cfg.StoreDirectory, unix.W_OK); err != nil {
		err = fmt.Errorf("directory '%s' is not writable", srv.cfg.StoreDirectory)
		logrus.Error(err)
		return err
	}
	srv.running = true
	return err
}

// Stop prevents new client capabilities from opening.
// Existing clients might need time to finish.
func (srv *StateService) Stop() error {
	logrus.Infof("stopping state store service")
	// build the list of stores to close to allow list update while closing
	srv.mux.Lock()
	if !srv.running {
		return fmt.Errorf("service already stopped")
	}
	srv.running = true
	clientList := make([]string, 0, len(srv.clientStores))
	for clientID, store := range srv.clientStores {
		if store != nil {
			clientList = append(clientList, clientID)
		}
	}
	if len(clientList) > 0 {
		logrus.Warningf("state store has stopped. %d client buckets are still running", len(clientList))
	} else {
		logrus.Infof("state store service has stopped properly. No clients remaining.")
	}
	srv.mux.Unlock()

	// Note that closing the store without releasing its clients can result in calls after
	// its store is running. The store should be able to handle this.
	//for _, clientID := range clientList {
	//	srv.mux.Lock()
	//	clientStore := srv.clientStores[clientID]
	//	srv.clientStores[clientID] = nil
	//	srv.mux.Unlock()
	//	if clientStore != nil {
	//		logrus.Infof("Stopping store for '%s'", clientID)
	//		clientStore.Close()
	//	}
	//}
	return nil
}

// NewStateStoreService creates a state storage server instance.
// Intended for use by the launcher.
//
//	storeDirectory is the location to store client state files
func NewStateStoreService(stateConfig config.StateConfig) *StateService {
	// TODO: use configuration to determine backend and load limits
	//kvStore := kvstore.NewKVStore(stateConfig.DatabaseURL)
	srv := &StateService{
		cfg:          stateConfig,
		clientStores: make(map[string]bucketstore.IBucketStore),
		clientRefs:   make(map[string]int),
		mux:          sync.Mutex{},
	}
	return srv
}
