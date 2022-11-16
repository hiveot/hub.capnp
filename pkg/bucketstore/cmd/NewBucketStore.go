package cmd

import (
	"path"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/bucketstore/bolts"
	"github.com/hiveot/hub/pkg/bucketstore/kvbtree"
	"github.com/hiveot/hub/pkg/bucketstore/pebble"
)

// NewBucketStore is a helper to create a new bucket store of a given type
//
//		directory is the directory to create the store
//	 clientID is used to name the store
//		backend is the type of store to create: BackendKVBTree, BackendBBolt, BackendPebble
func NewBucketStore(directory, clientID, backend string) (store bucketstore.IBucketStore) {
	if backend == bucketstore.BackendKVBTree {
		// kvbtree stores data into a single file
		storePath := path.Join(directory, clientID+".json")
		store = kvbtree.NewKVStore(clientID, storePath)
	} else if backend == bucketstore.BackendBBolt {
		// bbolt stores data into a single file
		storePath := path.Join(directory, clientID+".boltdb")
		store = bolts.NewBoltStore(clientID, storePath)
	} else if backend == bucketstore.BackendPebble {
		// Pebbles stores data into a directory
		storePath := path.Join(directory, clientID)
		store = pebble.NewPebbleStore(clientID, storePath)
	} else {
		logrus.Errorf("Unknown backend %s", backend)
	}
	return store
}
