package service

import (
	"context"

	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/history"
)

// GetPropertiesFunc is a callback function to retrieve latest properties of a Thing
// latest properties are stored separate from the history.
type GetPropertiesFunc func(thingAddr string, names []string) []*thing.ThingValue

// ReadHistory provides read access to the history of a thing
// This implements the IReadHistory interface
type ReadHistory struct {
	clientID string
	// routing address of the thing to read history of
	thingAddr string
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
	historyCursor := NewHistoryCursor(svc.thingAddr, name, cursor)
	return historyCursor
}

// GetProperties returns the most recent property and event values of the Thing
// Latest Properties are tracked in a 'latest' record which holds a map of propertyName:ThingValue records
//
//	providing 'names' can speed up read access significantly
func (svc *ReadHistory) GetProperties(_ context.Context, names []string) (values []*thing.ThingValue) {
	values = svc.getPropertiesFunc(svc.thingAddr, names)
	return values
}

// Info returns the history storage information of the thing
// availability of information depends on the underlying store to use.
func (svc *ReadHistory) Info(_ context.Context) *bucketstore.BucketStoreInfo {
	return svc.thingBucket.Info()
}

// Release closes the bucket
func (hc *ReadHistory) Release() {
	_ = hc.thingBucket.Close()
}

// NewReadHistory returns the capability to read from a thing's history
//
//	thingAddr is the Things's full address, usually publisherID/thingID
//	thingBucket is the bucket used to store history data
//	gePropertiesFunc implements the aggregation of the Thing's most recent property values
func NewReadHistory(clientID, thingAddr string, thingBucket bucketstore.IBucket, getPropertiesFunc GetPropertiesFunc) *ReadHistory {
	svc := &ReadHistory{
		clientID:          clientID,
		thingAddr:         thingAddr,
		thingBucket:       thingBucket,
		getPropertiesFunc: getPropertiesFunc,
	}
	return svc
}
