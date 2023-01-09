package service

import (
	"context"
	"encoding/json"

	"github.com/hiveot/hub.capnp/go/vocab"
	"github.com/hiveot/hub/lib/thing"

	"github.com/hiveot/hub/pkg/bucketstore/kvbtree"
	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryService is a wrapper around the internal bucket store
// This implements the IDirectory interface
type DirectoryService struct {
	store *kvbtree.KVBTreeStore
	// hubID is the gatewayID of this Hub, used when publishing Things of Hub services
	hubID        string
	serviceID    string // thingID of the service
	tdBucketName string
}

// CapReadDirectory provides the service to read the directory
func (srv *DirectoryService) CapReadDirectory(
	_ context.Context, clientID string) (directory.IReadDirectory, error) {
	bucket := srv.store.GetBucket(srv.tdBucketName)
	rd := NewReadDirectory(clientID, bucket)
	return rd, nil
}

// CapUpdateDirectory provides the service to update the directory
func (srv *DirectoryService) CapUpdateDirectory(
	_ context.Context, clientID string) (directory.IUpdateDirectory, error) {
	bucket := srv.store.GetBucket(srv.tdBucketName)
	ud := NewUpdateDirectory(clientID, bucket)
	return ud, nil
}

// Create a new Thing TD document describing this service
func (srv *DirectoryService) createServiceTD() *thing.TD {
	title := "Directory Store Service"
	deviceType := vocab.DeviceTypeService
	td := thing.NewTD(srv.serviceID, title, deviceType)

	return td
}

// Start opens the store and updates the service own TD
func (srv *DirectoryService) Start(ctx context.Context) error {
	err := srv.store.Open()
	if err == nil {
		myTD := srv.createServiceTD()
		myTDJSON, _ := json.Marshal(myTD)
		ud, err2 := srv.CapUpdateDirectory(ctx, directory.ServiceName)
		err = err2
		if err == nil {
			err = ud.UpdateTD(ctx, srv.hubID, myTD.ID, myTDJSON)
			ud.Release()
		}
	}
	return err
}

// Stop the storage server and flush changes to disk
func (srv *DirectoryService) Stop() error {
	err := srv.store.Close()
	return err
}

// NewDirectoryService creates a service to access TD documents
// This is using the KV bucket store.
//
//	thingStorePath is the file holding the directory data.
func NewDirectoryService(hubID string, thingStorePath string) *DirectoryService {

	kvStore := kvbtree.NewKVStore(directory.ServiceName, thingStorePath)
	svc := &DirectoryService{
		store:        kvStore,
		hubID:        hubID,
		tdBucketName: directory.TDBucketName,
		serviceID:    "urn:" + directory.ServiceName,
	}
	return svc
}
