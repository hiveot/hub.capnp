package service

import (
	"context"

	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/history"
)

const PropertiesBucketName = "properties"

// HistoryService provides storage for action and event history using the bucket store
// Each thingID has a bucket with actions and a bucket with events and actions.
// This implements the IHistoryService interface
type HistoryService struct {
	// The history service bucket store with a bucket for each thingID
	bucketStore bucketstore.IBucketStore
	propsStore  *PropertiesStore
}

// CapAddHistory provides the capability to update history
func (srv *HistoryService) CapAddHistory(_ context.Context, thingID string) history.IAddHistory {
	bucket := srv.bucketStore.GetBucket(thingID)
	historyUpdater := NewAddHistory(thingID, bucket, srv.propsStore.HandleAddValue)
	return historyUpdater
}

// CapAddAnyThing provides the capability to add to the history of any Thing.
// It is similar to CapAddHistory but not constrained to a specific Thing.
// This capability should only be provided to trusted services that capture events from multiple sources
// and can verify their authenticity.
func (srv *HistoryService) CapAddAnyThing(context.Context) history.IAddHistory {

	historyUpdater := NewAddAnyThing(srv.bucketStore, srv.propsStore.HandleAddValue)
	return historyUpdater
}

// CapReadHistory provides the capability to read history
func (srv *HistoryService) CapReadHistory(_ context.Context, thingID string) history.IReadHistory {
	bucket := srv.bucketStore.GetBucket(thingID)
	readHistory := NewReadHistory(thingID, bucket, srv.propsStore.GetProperties)
	//cursor := bucket.Cursor()
	//historyCursor := NewHistoryCursor(thingID, "", cursor)
	return readHistory
}

// GetValues provides the capability to read
//func (srv *HistoryService) GetValues(_ context.Context, thingNames []string) []thing.ThingValue {
//	bucket := srv.bucketStore.GetBucket(thingID)
//	cursor := bucket.Cursor()
//	historyCursor := NewHistoryCursor(thingID, "", cursor)
//	return historyCursor
//}

// Start using the history service
func (srv *HistoryService) Start(_ context.Context) error {
	return nil
}

// Stop using the history service and release resources
func (srv *HistoryService) Stop(_ context.Context) error {
	err := srv.propsStore.SaveChanges()
	return err
}

// NewHistoryService creates a new instance for the history service using the given
// storage bucket.
func NewHistoryService(store bucketstore.IBucketStore) *HistoryService {
	propsbucket := store.GetBucket(PropertiesBucketName)
	svc := &HistoryService{
		bucketStore: store,
		propsStore:  NewPropertiesStore(propsbucket),
	}
	return svc
}
