package service

import (
	"context"

	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/history"
)

const PropertiesBucketName = "properties"

// HistoryService provides storage for action and event history using the bucket store
// Each Thing has a bucket with events and actions.
// This implements the IHistoryService interface
type HistoryService struct {
	// The history service bucket store with a bucket for each Thing
	bucketStore bucketstore.IBucketStore
	propsStore  *PropertiesStore
	serviceID   string
}

// CapAddHistory provides the capability to update history
func (srv *HistoryService) CapAddHistory(
	_ context.Context, clientID string, publisherID, thingID string) (history.IAddHistory, error) {
	thingAddr := publisherID + "/" + thingID
	bucket := srv.bucketStore.GetBucket(thingAddr)

	historyUpdater := NewAddHistory(clientID, publisherID, thingID, bucket, srv.propsStore.HandleAddValue)
	return historyUpdater, nil
}

// CapAddAnyThing provides the capability to add to the history of any Thing.
// It is similar to CapAddHistory but not constrained to a specific Thing.
// This capability should only be provided to trusted services that capture events from multiple sources
// and can verify their authenticity.
func (srv *HistoryService) CapAddAnyThing(
	_ context.Context, clientID string) (history.IAddHistory, error) {

	historyUpdater := NewAddAnyThing(clientID, srv.bucketStore, srv.propsStore.HandleAddValue)
	return historyUpdater, nil
}

// CapReadHistory provides the capability to read history
func (srv *HistoryService) CapReadHistory(
	_ context.Context, clientID, publisherID, thingID string) (history.IReadHistory, error) {

	thingAddr := publisherID + "/" + thingID
	bucket := srv.bucketStore.GetBucket(thingAddr)
	readHistory := NewReadHistory(clientID, publisherID, thingID, bucket, srv.propsStore.GetProperties)
	return readHistory, nil
}

// Start using the history service
func (srv *HistoryService) Start(_ context.Context) error {
	return nil
}

// Stop using the history service and release resources
func (srv *HistoryService) Stop() error {
	err := srv.propsStore.SaveChanges()
	return err
}

// NewHistoryService creates a new instance for the history service using the given
// storage bucket.
//
//	store contains the bucket store to use
//	serviceID is the thingID of the service, eg "urn:history"
func NewHistoryService(store bucketstore.IBucketStore, serviceID string) *HistoryService {
	if serviceID == "" {
		serviceID = history.ServiceName
	}
	propsbucket := store.GetBucket(PropertiesBucketName)
	svc := &HistoryService{
		bucketStore: store,
		propsStore:  NewPropertiesStore(propsbucket),
		serviceID:   serviceID,
	}
	return svc
}
