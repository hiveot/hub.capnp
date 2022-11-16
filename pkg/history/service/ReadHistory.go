package service

import (
	"context"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/history"
)

type GetPropertiesFunc func(thingID string, names []string) []*thing.ThingValue

// ReadHistory provides read access to the history of a thing
// This implements the IReadHistory interface
type ReadHistory struct {
	// thing to read history of
	thingID string
	// The bucket containing the thing data
	thingBucket bucketstore.IBucket

	// The service implements the getPropertyValues function as it does the caching and
	// provides concurrency control.
	getPropertiesFunc GetPropertiesFunc
}

// GetEventHistory provides a cursor to iterate the event history of the thing
// name is used to filter on the event/action name. "" to iterate all events.
func (svc *ReadHistory) GetEventHistory(_ context.Context, name string) history.IHistoryCursor {
	cursor := svc.thingBucket.Cursor()
	historyCursor := NewHistoryCursor(svc.thingID, name, cursor)
	return historyCursor
}

// GetProperties returns the most recent property and event values of the Thing
// Latest Properties are tracked in a 'latest' record which holds a map of propertyName:ThingValue records
//
//	providing 'names' can speed up read access significantly
func (svc *ReadHistory) GetProperties(_ context.Context, names []string) (values []*thing.ThingValue) {
	values = svc.getPropertiesFunc(svc.thingID, names)
	return values
}

// Info returns the history storage information of the thing
// availability of information depends on the underlying store to use.
func (svc *ReadHistory) Info(_ context.Context) *bucketstore.BucketStoreInfo {
	return svc.thingBucket.Info()
}

// Release closes the bucket
func (hc *ReadHistory) Release() {
	hc.thingBucket.Close()
}

// NewReadHistory returns the capability to read from a thing's history
//
//	thingID is the Things's full ID
//	thingBucket is the bucket used to store history data
//	gePropertiesFunc implements the aggregation of the Thing's most recent property values
func NewReadHistory(thingID string, thingBucket bucketstore.IBucket, getPropertiesFunc GetPropertiesFunc) *ReadHistory {
	svc := &ReadHistory{
		thingID:           thingID,
		thingBucket:       thingBucket,
		getPropertiesFunc: getPropertiesFunc,
	}
	return svc
}
