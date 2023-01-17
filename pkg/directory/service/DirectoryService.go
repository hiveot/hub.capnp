package service

import (
	"context"
	"encoding/json"

	"github.com/hiveot/hub.capnp/go/vocab"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/pubsub"

	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryService is a wrapper around the internal bucket store
// This implements the IDirectory interface
type DirectoryService struct {
	servicePubSub pubsub.IServicePubSub
	store         bucketstore.IBucketStore
	serviceID     string // thingID of the service instance
	tdBucketName  string
}

// CapReadDirectory provides the service to read the directory
func (svc *DirectoryService) CapReadDirectory(
	_ context.Context, clientID string) (directory.IReadDirectory, error) {
	bucket := svc.store.GetBucket(svc.tdBucketName)
	rd := NewReadDirectory(clientID, bucket)
	return rd, nil
}

// CapUpdateDirectory provides the service to update the directory
func (svc *DirectoryService) CapUpdateDirectory(
	_ context.Context, clientID string) (directory.IUpdateDirectory, error) {
	bucket := svc.store.GetBucket(svc.tdBucketName)
	ud := NewUpdateDirectory(clientID, bucket)
	return ud, nil
}

// Create a new Thing TD document describing this service
func (svc *DirectoryService) createServiceTD() *thing.TD {
	title := "Directory Store Service"
	deviceType := vocab.DeviceTypeService
	td := thing.NewTD(svc.serviceID, title, deviceType)

	return td
}

func (svc *DirectoryService) handleTDEvent(event *thing.ThingValue) {
	ctx := context.Background()
	// TODO: reserve a capability for this instead of create/release
	ud, err := svc.CapUpdateDirectory(ctx, directory.ServiceName)
	if err == nil {
		err = ud.UpdateTD(ctx, event.PublisherID, event.ThingID, event.ValueJSON)
		ud.Release()
	}
}

// Start opens the store and publishes the service's own TD
func (svc *DirectoryService) Start(ctx context.Context) error {
	err := svc.store.Open()

	// subscribe to TD events to add to the directory
	if err == nil && svc.servicePubSub != nil {
		err = svc.servicePubSub.SubTDs(ctx, svc.handleTDEvent)
	}

	if err == nil {
		myTD := svc.createServiceTD()
		myTDJSON, _ := json.Marshal(myTD)
		if svc.servicePubSub != nil {
			// publish the TD
			err = svc.servicePubSub.PubTD(ctx, svc.serviceID, myTD.DeviceType, myTDJSON)
		} else {
			// no pubsub, so store the TD
			ud, err2 := svc.CapUpdateDirectory(ctx, directory.ServiceName)
			err = err2
			if err == nil {
				err = ud.UpdateTD(ctx, svc.serviceID, myTD.ID, myTDJSON)
				ud.Release()
			}
		}
	}

	return err
}

// Stop the storage server and flush changes to disk
func (svc *DirectoryService) Stop() error {
	if svc.servicePubSub != nil {
		svc.servicePubSub.Release()
	}
	err := svc.store.Close()
	return err
}

// NewDirectoryService creates a service to access TD documents
// This is using the KV bucket store.
//
//	serviceID is the instance ID of the service. Default ("") is the directory service name
//	store bucket store for persisting  the directory data. This will be opened on Start and closed on Stop.
//	servicePubSub is the pubsub to use to subscribe to directory events. Will be released on Stop.
func NewDirectoryService(
	serviceID string, store bucketstore.IBucketStore, servicePubSub pubsub.IServicePubSub) *DirectoryService {
	if serviceID == "" {
		serviceID = directory.ServiceName
	}
	//kvStore := kvbtree.NewKVStore(serviceID, thingStorePath)
	svc := &DirectoryService{
		servicePubSub: servicePubSub,
		store:         store,
		serviceID:     serviceID,
		tdBucketName:  directory.TDBucketName,
	}
	return svc
}
