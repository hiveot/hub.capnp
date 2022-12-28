package service

import (
	"context"
	"strconv"
	"time"

	"github.com/araddon/dateparse"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub.go/pkg/vocab"
	"github.com/hiveot/hub/pkg/bucketstore"
)

// AddAnyThing adds events and actions of any Thing
// this is not restricted to one Thing and only intended for services that are authorized to do so.
type AddAnyThing struct {
	clientID string
	// store with buckets for Things
	store bucketstore.IBucketStore
	// onAddedValue is a callback to invoke after a value is added. Intended for tracking most recent values.
	onAddedValue func(ev *thing.ThingValue, isAction bool)
}

// encode a ThingValue into a single key value pair
// Encoding generates a key as: timestampMsec/name/a|e, where a|e indicates action or event
func (svc *AddAnyThing) encodeValue(thingValue *thing.ThingValue, isAction bool) (key string, val []byte) {
	var err error
	ts := time.Now()
	if thingValue.Created != "" {
		ts, err = dateparse.ParseAny(thingValue.Created)
		if err != nil {
			logrus.Infof("Invalid Created time '%s'. Using current time instead", thingValue.Created)
			ts = time.Now()
		}
	}

	// the index uses milliseconds for timestamp
	timestamp := ts.UnixMilli()
	key = strconv.FormatInt(timestamp, 10) + "/" + thingValue.Name
	if isAction {
		key = key + "/a"
	} else {
		key = key + "/e"
	}
	// TODO: reorganize data to store. Remove duplication. Timestamp in msec since epoc
	val = thingValue.ValueJSON
	return key, val
}

// AddAction adds a Thing action with the given name and value to the action history
// value is json encoded. Optionally include a 'created' ISO8601 timestamp
func (svc *AddAnyThing) AddAction(_ context.Context, actionValue *thing.ThingValue) error {
	key, val := svc.encodeValue(actionValue, true)
	bucket := svc.store.GetBucket(actionValue.ThingAddr)
	err := bucket.Set(key, val)
	_ = bucket.Close()
	if svc.onAddedValue != nil {
		svc.onAddedValue(actionValue, true)
	}
	return err
}

// AddEvent adds an event to the event history
// If the event has no created time, it will be set to 'now'
func (svc *AddAnyThing) AddEvent(_ context.Context, eventValue *thing.ThingValue) error {
	if eventValue.Created == "" {
		eventValue.Created = time.Now().Format(vocab.ISO8601Format)
	}
	key, val := svc.encodeValue(eventValue, false)
	bucket := svc.store.GetBucket(eventValue.ThingAddr)
	err := bucket.Set(key, val)
	_ = bucket.Close()
	if svc.onAddedValue != nil {
		svc.onAddedValue(eventValue, false)
	}
	return err
}

// AddEvents provides a bulk-add of events to the event history
// If this service is constraint to a thing then reject requests with wrong thing address
func (svc *AddAnyThing) AddEvents(ctx context.Context, eventValues []*thing.ThingValue) (err error) {
	if eventValues == nil || len(eventValues) == 0 {
		return nil
	} else if len(eventValues) == 1 {
		err = svc.AddEvent(ctx, eventValues[0])
		return err
	}
	// encode events as K,V pair and group them by thingAddr
	kvpairsByThingAddr := make(map[string]map[string][]byte)
	for _, eventValue := range eventValues {
		// kvpairs hold a map of storage encoded value key and value
		kvpairs, found := kvpairsByThingAddr[eventValue.ThingAddr]
		if !found {
			kvpairs = make(map[string][]byte, 0)
			kvpairsByThingAddr[eventValue.ThingAddr] = kvpairs
		}
		key, value := svc.encodeValue(eventValue, false)
		kvpairs[key] = value
		// notify owner to update thing properties
		if svc.onAddedValue != nil {
			svc.onAddedValue(eventValue, false)
		}
	}
	// adding in bulk, opening and closing buckets only once for each thing address
	for thingAddr, kvpairs := range kvpairsByThingAddr {
		bucket := svc.store.GetBucket(thingAddr)
		_ = bucket.SetMultiple(kvpairs)
		err = bucket.Close()
	}
	return nil
}

// Release the capability and its resources
func (svc *AddAnyThing) Release() {

}

// NewAddAnyThing provides the capability to add values to Thing history buckets
// onAddedValue is invoked after the value is added to the bucket.
func NewAddAnyThing(
	clientID string,
	store bucketstore.IBucketStore,
	onAddedValue func(value *thing.ThingValue, isAction bool)) *AddAnyThing {
	svc := &AddAnyThing{
		clientID:     clientID,
		store:        store,
		onAddedValue: onAddedValue,
	}

	return svc
}
