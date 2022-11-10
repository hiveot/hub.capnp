package service

import (
	"context"
	"encoding/json"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub.go/pkg/vocab"
	"github.com/hiveot/hub/internal/bucketstore/kvmem"
	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryService is a wrapper around the internal bucket store
// This implements the IDirectory interface
type DirectoryService struct {
	store        *kvmem.KVMemStore
	tdBucketName string
}

// Create a new Thing TD document describing this service
func (srv *DirectoryService) createServiceTD() *thing.ThingDescription {
	thingID := thing.CreateThingID("", directory.ServiceName, vocab.DeviceTypeService)
	title := "Directory Store Service"
	deviceType := vocab.DeviceTypeService
	td := thing.CreateTD(thingID, title, deviceType)

	return td
}

// CapReadDirectory provides the service to read the directory
func (srv *DirectoryService) CapReadDirectory(ctx context.Context) directory.IReadDirectory {
	bucket := srv.store.GetBucket(srv.tdBucketName)
	rd := NewReadDirectory(bucket)
	return rd
}

// CapUpdateDirectory provides the service to update the directory
func (srv *DirectoryService) CapUpdateDirectory(ctx context.Context) directory.IUpdateDirectory {
	bucket := srv.store.GetBucket(srv.tdBucketName)
	ud := NewUpdateDirectory(bucket)
	return ud
}

// Start opens the store and updates the service own TD
func (srv *DirectoryService) Start(ctx context.Context) error {
	err := srv.store.Open()
	if err == nil {
		myTD := srv.createServiceTD()
		myTDJson, _ := json.Marshal(myTD)
		ud := srv.CapUpdateDirectory(ctx)
		err = ud.UpdateTD(ctx, myTD.ID, string(myTDJson))
		ud.Release()
	}
	return err
}

// Stop the storage server and flush changes to disk
func (srv *DirectoryService) Stop(ctx context.Context) error {
	_ = ctx
	err := srv.store.Close()
	return err
}

// NewDirectoryService creates a service to access TD documents
//
//	thingStorePath is the file holding the directory data.
func NewDirectoryService(ctx context.Context, thingStorePath string) *DirectoryService {

	kvStore := kvmem.NewKVStore(directory.ServiceName, thingStorePath)
	svc := &DirectoryService{
		store:        kvStore,
		tdBucketName: directory.TDBucketName,
	}
	return svc
}
