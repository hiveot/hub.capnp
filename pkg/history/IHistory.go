// Package history with POGS definitions of the history store.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package history

import (
	"context"

	"github.com/hiveot/hub.go/pkg/thing"
)

//
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
//	created string
//}

// StoreInfo with history store information
type StoreInfo struct {

	// Storage engine used, eg "mongodb" or other
	Engine string

	// The number of actions in the store
	NrActions int

	// The number of events in the store
	NrEvents int

	// Nr of seconds the service is running
	Uptime int
}

// IHistory defines a POGS based capability API of the thing history
type IHistory interface {
	// CapReadHistory provides the capability to read history
	CapReadHistory() IReadHistory

	// CapUpdateHistory provides the capability to update history
	CapUpdateHistory() IUpdateHistory
}

// IReadHistory defines the POGS based capability to read Thing history
type IReadHistory interface {

	// GetActionHistory returns the history of a Thing action
	// before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)
	GetActionHistory(ctx context.Context, thingID string, actionName string, after string, before string, limit int) (values []thing.ThingValue, err error)

	// GetEventHistory returns the history of a Thing event
	// before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)
	GetEventHistory(ctx context.Context, thingID string, eventName string, after string, before string, limit int) (values []thing.ThingValue, err error)

	// GetLatestEvents returns a map of the latest event values of a Thing
	GetLatestEvents(ctx context.Context, thingID string) (latest map[string]thing.ThingValue, err error)

	// Info return storage information
	Info(ctx context.Context) (info StoreInfo, err error)
}

// IUpdateHistory defines the POGS based capability to update the Thing history
type IUpdateHistory interface {

	// AddAction adds a Thing action with the given name and value to the action history
	// value is json encoded. Optionally include a 'created' ISO8601 timestamp
	AddAction(ctx context.Context, actionValue thing.ThingValue) error

	// AddEvent adds an event to the event history
	AddEvent(ctx context.Context, eventValue thing.ThingValue) error

	// AddEvents provides a bulk-add of events to the event history
	AddEvents(ctx context.Context, eventValues []thing.ThingValue) error
}
