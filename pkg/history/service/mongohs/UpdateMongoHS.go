// Package mongohs with MongoDB based history store
// This implements the HistoryStore.proto API
package mongohs

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/hiveot/hub.go/pkg/thing"
)

// AddAction adds a new action to the history store
func (srv *MongoHistoryServer) AddAction(ctx context.Context,
	actionValue thing.ThingValue) error {

	// Name and ThingID are required fields
	if actionValue.Name == "" || actionValue.ThingID == "" {
		err := fmt.Errorf("missing name or thingID")
		logrus.Warning(err)
		return err
	}
	if actionValue.Created == "" {
		actionValue.Created = time.Now().UTC().Format(time.RFC3339)
	}

	// It would be nice to simply use bson marshal, but that isn't possible as the
	// required timestamp needs to be added in BSON format.
	createdTime, err := time.Parse(time.RFC3339, actionValue.Created)
	timestamp := primitive.NewDateTimeFromTime(createdTime)
	evBson := bson.M{
		TimeStampField: timestamp,
		"metadata":     bson.M{"thingID": actionValue.ThingID, "name": actionValue.Name},
		"name":         actionValue.Name,
		"thingID":      actionValue.ThingID,
		"value":        actionValue.ValueJSON,
		"created":      actionValue.Created,
	}
	res, err := srv.actionCollection.InsertOne(ctx, evBson)
	_ = res
	return err
}

// AddEvent adds a new event to the history store
// The event 'created' field will be used as timestamp after parsing it using time.RFC3339
func (srv *MongoHistoryServer) AddEvent(
	ctx context.Context, eventValue thing.ThingValue) error {

	// Name and ThingID are required fields
	if eventValue.Name == "" || eventValue.ThingID == "" {
		err := fmt.Errorf("missing name or thingID")
		logrus.Warning(err)
		return err
	}
	if eventValue.Created == "" {
		eventValue.Created = time.Now().UTC().Format(time.RFC3339)
	}

	// It would be nice to simply use bson marshal, but that isn't possible as the
	// required timestamp needs to be added in BSON format.
	//createdTime, err := time.Parse("2006-01-02T15:04:05-07:00", event.Created)
	createdTime, err := time.Parse(time.RFC3339, eventValue.Created)
	timestamp := primitive.NewDateTimeFromTime(createdTime)
	evBson := bson.M{
		TimeStampField: timestamp,
		//"metadata":     bson.M{"thingID": event.ThingID},
		//"metadata": bson.M{"thingID": thingID, "name": name},
		"metadata": bson.M{"name": eventValue.Name},
		"name":     eventValue.Name,
		"thingID":  eventValue.ThingID,
		"value":    eventValue.ValueJSON,
		"created":  eventValue.Created,
	}

	// TODO: support different granularity by using expireAfterSeconds
	// although without downsampling this might not be useful
	res, err := srv.eventCollection.InsertOne(ctx, evBson)
	_ = res
	if err != nil {
		logrus.Error(err)
		return err
	}
	//return nil, nil

	// Last, track the event value in the 'latest' collection of the Thing properties.
	// This collection has a row per thingID with properties for each of the event names.
	// Unfortunately this doubles the duration of AddEvent :(
	if srv.useSeparateLatestTable {
		err = srv.addLatest(ctx, eventValue)
	}
	//--- end test 2
	if err != nil {
		logrus.Error(err)
	}
	return err
}

// AddEvents performs a bulk update of events
// This provides a significant performance increase over adding multiple single events
// The event 'created' field will be used as timestamp after parsing it using time.RFC3339
func (srv *MongoHistoryServer) AddEvents(ctx context.Context,
	events []thing.ThingValue) error {
	evList := make([]interface{}, 0)

	// convert to an array of bson objects
	for _, event := range events {

		// Name and ThingID are required fields
		if event.Name == "" || event.ThingID == "" {
			err := fmt.Errorf("missing name or thingID")
			logrus.Warning(err)
			return err
		}
		if event.Created == "" {
			event.Created = time.Now().UTC().Format(time.RFC3339)
		}

		// It would be nice to simply use bson marshal, but that isn't possible as the
		// required timestamp needs to be added in BSON format.
		//createdTime, err := time.Parse("2006-01-02T15:04:05-07:00", event.Created)
		createdTime, _ := time.Parse(time.RFC3339, event.Created)
		timestamp := primitive.NewDateTimeFromTime(createdTime)
		evBson := bson.M{
			TimeStampField: timestamp,
			//"metadata":     bson.M{"thingID": event.ThingID},
			//"metadata": bson.M{"thingID": event.ThingID, "name": event.Name},
			"metadata": bson.M{"name": event.Name},
			"name":     event.Name,
			"thingID":  event.ThingID,
			"value":    event.ValueJSON,
			"created":  event.Created,
		}
		evList = append(evList, evBson)
	}
	// TODO: support different granularity by using expireAfterSeconds
	// although without downsampling this might not be useful
	res, err := srv.eventCollection.InsertMany(ctx, evList)
	_ = res

	//

	// Last, track the event value in the 'latest' collection of the Thing properties.
	// This collection has a row per thingID with properties for each of the event names.
	// Unfortunately this doubles the duration of AddEvent :(
	if srv.useSeparateLatestTable {
		// reduce all samples to those with the highest timestamp for the given thing and value name
		// build our own latest document for each thingID before updating
		latestThings := make(map[string]map[string]thing.ThingValue)
		for _, event := range events {
			// map of values by thing ID
			thingValues, found := latestThings[event.ThingID]
			if found {
				// if the value has a sensor of the same name
				value, found := thingValues[event.Name]
				if found {
					if value.Created < event.Created {
						// replace it if it is older than the event to update
						thingValues[event.Name] = event
					}
				} else {
					// the thing does not yet have the value
					thingValues[event.Name] = event
				}
			} else {
				// this thing is new
				thingValues = make(map[string]thing.ThingValue)
				thingValues[event.Name] = event
				latestThings[event.ThingID] = thingValues
			}
		}
		// Next update each thing
		for _, newValues := range latestThings {
			for _, value := range newValues {
				err = srv.addLatest(ctx, value)
				if err != nil {
					logrus.Error(err)
					return err
				}
			}
		}
	}
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}

// Update the latest record with the new value
func (srv *MongoHistoryServer) addLatest(ctx context.Context,
	value thing.ThingValue) error {

	// It is possible that events arrive out of order so the created date must be newer
	// than the existing date
	filter := bson.D{
		{"thingID", value.ThingID},
	}
	// Translation of the following pipeline:
	// if event.Created > {document}[event.Name].created {
	//    {document}[event.Name] = event
	// }
	pipeline := bson.A{
		bson.M{"$set": bson.M{value.Name: bson.M{"$cond": bson.A{
			bson.M{"$gt": bson.A{
				value.Created, "$" + value.Name + ".created",
			}},
			value,            // replace with new value
			"$" + value.Name, // or keep existing
		}}}},
	}

	//pipeline := bson.M{"$set": bson.M{event.Name: event}}
	opts := options.UpdateOptions{}

	opts.SetUpsert(true)
	//opts.SetHint("thingID")
	res2, err := srv.latestEvents.UpdateOne(ctx, filter, pipeline, &opts)
	_ = res2
	return err
}
