package service

import (
	"context"
	"encoding/json"
	"github.com/hiveot/hub/api/go/hubapi"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/api/go/vocab"
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

	logrus.Infof("clientID=%s", clientID)
	bucket := svc.store.GetBucket(svc.tdBucketName)
	rd := NewReadDirectory(clientID, bucket)
	return rd, nil
}

// CapUpdateDirectory provides the service to update the directory
func (svc *DirectoryService) CapUpdateDirectory(
	_ context.Context, clientID string) (directory.IUpdateDirectory, error) {
	logrus.Infof("clientID=%s", clientID)
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

func (svc *DirectoryService) handleTDEvent(event thing.ThingValue) {
	ctx := context.Background()
	// TODO: reserve a capability for this instead of create/release
	ud, err := svc.CapUpdateDirectory(ctx, directory.ServiceName)
	if err == nil {
		err = ud.UpdateTD(ctx, event.PublisherID, event.ThingID, event.Data)
		ud.Release()
	}
}

// Start the directory service and publish the service's own TD
// This subscribes to pubsub TD events and updates the directory.
func (svc *DirectoryService) Start() (err error) {
	ctx := context.Background()

	// subscribe to TD events to add to the directory
	if svc.servicePubSub != nil {
		err = svc.servicePubSub.SubEvent(ctx, "", "", hubapi.EventNameTD, svc.handleTDEvent)
	}

	if err == nil {
		myTD := svc.createServiceTD()
		myTDJSON, _ := json.Marshal(myTD)
		if svc.servicePubSub != nil {
			// publish the TD
			err = svc.servicePubSub.PubEvent(ctx, svc.serviceID, hubapi.EventNameTD, myTDJSON)
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

// Stop the service
func (svc *DirectoryService) Stop() error {
	if svc.servicePubSub != nil {
		svc.servicePubSub.Release()
	}
	return nil
}

func (svc *DirectoryService) Release() {}

// NewDirectoryService creates a service to access TD documents
// The servicePubSub is optional and ignored when nil. It is used to subscribe to directory events and
// will be released on Stop.
//
//	serviceID is the instance ID of the service. Default ("") is the directory service name
//	store is an open bucket store for persisting the directory data.
//	servicePubSub is the pubsub service
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
