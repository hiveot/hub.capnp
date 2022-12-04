package service

import (
	"context"
	"encoding/json"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub.go/pkg/vocab"
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

// Create a new Thing TD document describing this service
func (srv *DirectoryService) createServiceTD() *thing.ThingDescription {
	title := "Directory Store Service"
	deviceType := vocab.DeviceTypeService
	td := thing.CreateTD(srv.serviceID, title, deviceType)

	return td
}

// CapReadDirectory provides the service to read the directory
func (srv *DirectoryService) CapReadDirectory(_ context.Context) directory.IReadDirectory {
	bucket := srv.store.GetBucket(srv.tdBucketName)
	rd := NewReadDirectory(bucket)
	return rd
}

// CapUpdateDirectory provides the service to update the directory
func (srv *DirectoryService) CapUpdateDirectory(_ context.Context) directory.IUpdateDirectory {
	bucket := srv.store.GetBucket(srv.tdBucketName)
	ud := NewUpdateDirectory(bucket)
	return ud
}

// Start opens the store and updates the service own TD
func (srv *DirectoryService) Start(ctx context.Context) error {
	err := srv.store.Open()
	if err == nil {
		myTD := srv.createServiceTD()
		myTDJSON, _ := json.Marshal(myTD)
		myTDAddr := srv.hubID + "/" + myTD.ID
		ud := srv.CapUpdateDirectory(ctx)
		err = ud.UpdateTD(ctx, myTDAddr, myTDJSON)
		ud.Release()
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
