// Package mongohs with MongoDB based history store
// This implements the HistoryStore.proto API
package mongohs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/history"
)

// getHistory returns the request history from a collection
// if before is used, after must be set as well
func (srv *MongoHistoryServer) getHistory(ctx context.Context,
	collection *mongo.Collection,
	thingID string, valueName string, after string, before string, limit int) ([]thing.ThingValue, error) {

	var hist = make([]thing.ThingValue, 0)
	var timeFilter bson.D
	if collection == nil {
		err := fmt.Errorf("parameter error. Collection is nil")
		logrus.Error(err)
		return hist, err
	}

	filter := bson.M{}
	if thingID != "" {
		filter["thingID"] = thingID
	}
	// filter on a time range. Require at least an 'after' time.
	if before != "" && after == "" {
		err := fmt.Errorf("in a time range query before time requires after time to be provided")
		logrus.Warning(err)
		return nil, err
	}
	if after != "" {
		timeAfter, err := dateparse.ParseAny(after)
		if err != nil {
			logrus.Infof("Invalid 'After' time: %s", err)
			return nil, err
		}
		timeAfterBson := primitive.NewDateTimeFromTime(timeAfter)
		if before == "" {
			// not a range, just time after
			timeFilter = bson.D{{"$gte", timeAfterBson}}
		} else {
			// make it a range
			timeBefore, err := dateparse.ParseAny(before)
			if err != nil {
				logrus.Infof("Invalid 'Before' time: %s", err)
				return nil, err
			}
			timeBeforeBson := primitive.NewDateTimeFromTime(timeBefore)
			timeFilter = bson.D{{"$gte", timeAfterBson}, {"$lte", timeBeforeBson}}
		}
		filter[TimeStampField] = timeFilter
	}

	if valueName != "" {
		filter["name"] = valueName
	}
	//if limit > 0 {
	//	filter["limit"] = 0
	//}

	cursor, err := collection.Find(ctx, filter, options.Find().SetLimit(int64(limit)))
	if err != nil {
		logrus.Warning(err)
		return nil, err
	}

	defer cursor.Close(ctx)
	//res := make([]thing.ThingValue,0) &thing.ThingValueList{
	//	Values: actions,
	//}
	for cursor.Next(ctx) {
		histValue := thing.ThingValue{}
		err = cursor.Decode(&histValue)
		hist = append(hist, histValue)
	}
	return hist, err
}

// getLatestValuesFromTimeSeries using aggregate pipeline
// NOTE: THIS DOESN'T SCALE. 1 million records, 100 things, 10 sensor names
// takes a whopping 10 seconds to complete.
func (srv *MongoHistoryServer) getLatestValuesFromTimeSeries(
	ctx context.Context, thingID string) (map[string]thing.ThingValue, error) {

	values := make(map[string]thing.ThingValue)
	// equivalent to
	// db.events.aggregate([
	//   { $match: { "thingID": "thing-0" } },
	//   { $sort: { "timestamp": -1 } },
	//   { $group: { _id: "$metadata.name",
	//              name: { $first: "$name" },
	//              created: { $first: "$created" },
	//              value:{$first:"$value"},
	//             ]).explain("executionStats")
	//]).explain("executionStats")

	matchStage := bson.D{
		{"$match",
			bson.D{
				{"thingID", thingID},
			},
		},
	}
	sortStage := bson.D{
		{"$sort",
			bson.D{
				{"metadata.name", 1},
				{"timestamp", -1},
			},
		},
	}
	// grouping doesn't take advantage of sorted sequences
	// see: https://jira.mongodb.org/browse/SERVER-4507
	groupStage := bson.D{
		{"$group",
			bson.D{
				//with an index on metadata.name this should use DISTINCT_SCAN and be faster
				//https://www.mongodb.com/docs/v6.0/core/timeseries/timeseries-secondary-index/
				// However, this fails with a bug: memory usage for BoundedSorter is invalid error
				// https://jira.mongodb.org/browse/SERVER-68196
				{"_id", "$metadata.name"},
				//{"_id", "$name"},
				{"timestamp", bson.M{"$first": "$timestamp"}},
				{"name", bson.M{"$first": "$name"}},
				{"created", bson.M{"$first": "$created"}},
				{"value", bson.M{"$first": "$value"}},
			},
		},
	}
	pipeline := mongo.Pipeline{matchStage, sortStage, groupStage}
	aggOptions := &options.AggregateOptions{}
	cursor, err := srv.eventCollection.Aggregate(ctx, pipeline, aggOptions)
	defer cursor.Close(ctx)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	count := 0
	for cursor.Next(ctx) {
		var r1 map[string]interface{}
		err = cursor.Decode(&r1)

		value := thing.ThingValue{}
		// for a small number of results using FindOne to get the event details is faster,
		// but for a large number of results grouping is faster

		//filter1 := bson.M{"_id": r1["objectID"]}
		//one := srv.eventCollection.FindOne(ctx, filter1)
		//one.Decode(&value)
		err = cursor.Decode(&value)
		if err == nil {
			values[value.Name] = value
			count++
		}
	}
	return values, nil
}

// getLatestValuesFromCollection using a separate collection to get the latest
// This is very fast on read but doubles the write time :(
func (srv *MongoHistoryServer) getLatestValuesFromCollection(
	ctx context.Context, thingID string) (map[string]thing.ThingValue, error) {

	propValues := map[string]thing.ThingValue{}

	filter := bson.M{"thingID": thingID}
	res := srv.latestEvents.FindOne(ctx, filter)

	var thingValues map[string]interface{}
	err := res.Decode(&thingValues)
	if err != nil {
		return propValues, err
	}
	// ugly but otherwise unmarshal fails
	delete(thingValues, "_id")
	delete(thingValues, "thingID")

	asJson, err := json.Marshal(thingValues)
	err = json.Unmarshal(asJson, &propValues)
	return propValues, err
}

// GetLatestEvents returns the last received events of a Thing
func (srv *MongoHistoryServer) GetLatestEvents(ctx context.Context,
	thingID string) (map[string]thing.ThingValue, error) {
	var propValues map[string]thing.ThingValue
	var err error

	if srv.useSeparateLatestTable {
		propValues, err = srv.getLatestValuesFromCollection(ctx, thingID)
	} else {
		propValues, err = srv.getLatestValuesFromTimeSeries(ctx, thingID)
	}
	logrus.Infof("found %d different event names", len(propValues))

	return propValues, err
}

// Info returns store statistics
func (srv *MongoHistoryServer) Info(ctx context.Context) (info history.StoreInfo, err error) {
	nrActions, err := srv.actionCollection.CountDocuments(ctx, bson.D{})
	nrEvents, _ := srv.eventCollection.CountDocuments(ctx, bson.D{})
	uptime := time.Now().Sub(srv.startTime).Seconds()

	info = history.StoreInfo{
		Engine:    "mongodb",
		NrActions: int(nrActions),
		NrEvents:  int(nrEvents),
		Uptime:    int(uptime),
	}
	return info, err
}
