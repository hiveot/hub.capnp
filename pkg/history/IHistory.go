// Package history with POGS definitions of the history store.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package history

import (
	"context"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/bucketstore"
)

// ServiceName is the name of this service to connect to
const ServiceName = "historystore"

// EventNameProperties 'properties' is the name of the event that holds a JSON encoded map
// with one or more property values of a thing.
// TODO: link this to the definitions for submitting events
const EventNameProperties = "properties"

//// ThingValue containing an event or action value of a thing
//type ThingValue struct {
//
//	// Name of event or action as described in the thing TD
//	name string
//
//	// Value, JSON encoded
//	valueJSON string
//
//	// Timestamp the value was created, in ISO8601 format. Default is 'now'
//  //  YYYY-MM-DDTHH:MM:SS.sss-0900  = 28 char
//	created string
//}
//
//// StoreInfo with history store information
//type StoreInfo struct {
//
//	// Storage engine used, eg "mongodb" or other
//	Engine string
//
//	// The number of actions in the store
//	NrActions int
//
//	// The number of events in the store
//	NrEvents int
//
//	// Nr of seconds the service is running
//	Uptime int
//}

// IHistoryService defines the  capability to access the thing history service
type IHistoryService interface {

	// CapAddHistory provides the capability to add to the history of a Thing.
	// This capability should only be provided to the device or service that have write access to the Thing.
	//  thingAddr is the gateway address of the thing, eg publisherID/thingID
	CapAddHistory(ctx context.Context, thingAddr string) IAddHistory

	// CapAddAnyThing provides the capability to add to the history of any Thing.
	// It is similar to CapAddHistory but not constraint to a specific Thing.
	// This capability should only be provided to trusted services that capture events from multiple sources
	// and can verify their authenticity.
	CapAddAnyThing(ctx context.Context) IAddHistory

	// CapReadHistory provides the capability to iterate history.
	// This returns an iterator for the history.
	// Values added after creating the cursor might not be included, depending on the
	// underlying store.
	// This capability can be provided to anyone who has read access to the thing.
	//
	//  thingAddr is the gateway address of the thing, eg publisherID/thingID
	CapReadHistory(ctx context.Context, thingAddr string) IReadHistory

	// TBD: Subscribe to the pubsub service to receive events and actions
	//Subscribe()

	// Release the client
	//Release()
}

// IAddHistory defines the capability to add to a Thing's history
// If this capability was created with the thingAddr constraint then only values for this
// thingAddr will be accepted.
type IAddHistory interface {

	// AddAction adds a Thing action with the given name and value to the action history
	// The given value object must not be modified after this call.
	AddAction(ctx context.Context, thingValue *thing.ThingValue) error

	// AddEvent adds an event to the event history
	// The given value object must not be modified after this call.
	AddEvent(ctx context.Context, thingValue *thing.ThingValue) error

	// AddEvents provides a bulk-add of events to the event history
	// The given value objects must not be modified after this call.
	AddEvents(ctx context.Context, eventValues []*thing.ThingValue) error

	// Release the capability and its resources
	Release()
}

// IReadHistory defines the capability to read information from a thing
type IReadHistory interface {
	// GetEventHistory returns a cursor to iterate the history of the thing
	// name is the event or action to filter on. Use "" to iterate all events/action of the thing
	// The cursor MUST be released after use.
	GetEventHistory(ctx context.Context, name string) IHistoryCursor

	// GetProperties returns the most recent property and event values of the Thing
	//  names is the list of properties to return. Use nil or empty list to return all known properties.
	//  This returns a list of thing values.
	GetProperties(ctx context.Context, names []string) []*thing.ThingValue

	// Info returns the history storage information of the thing
	Info(ctx context.Context) *bucketstore.BucketStoreInfo

	// Release the capability and its resources
	Release()
}

// IHistoryCursor is a cursor to iterate the Thing event and action history
// Use Seek to find the start of the range and NextN to read batches of values
type IHistoryCursor interface {
	// First return the oldest value in the history
	// Returns nil if the store is empty
	First() (thingValue *thing.ThingValue, valid bool)

	// Last returns the latest value in the history
	// Returns nil if the store is empty
	Last() (thingValue *thing.ThingValue, valid bool)

	// Next returns the next value in the history
	// Returns nil when trying to read past the last value
	Next() (thingValue *thing.ThingValue, valid bool)

	// NextN returns a batch of next history values
	// Returns empty list when trying to read past the last value
	// itemsRemaining is true as long as more items can be retrieved
	NextN(steps uint) (batch []*thing.ThingValue, itemsRemaining bool)

	// Prev returns the previous value in history
	// Returns nil when trying to read before the first value
	Prev() (thingValue *thing.ThingValue, valid bool)

	// PrevN returns a batch of previous history values
	// Returns empty list when trying to read before the first value
	// itemsRemaining is true as long as more items can be retrieved
	PrevN(steps uint) (batch []*thing.ThingValue, itemsRemaining bool)

	// Release the cursor and resources
	Release()

	// Seek the starting point for iterating the history
	// This returns the value at timestamp or next closest if it doesn't exist
	// Returns empty list when there are no values at or past the given timestamp
	Seek(isoTimestamp string) (thingValue *thing.ThingValue, valid bool)
}
