// Package mongohs with MongoDB based history store
// This implements the HistoryStore.proto API
package mongohs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/config"
)

const TimeStampField = "timestamp"
const DefaultStoreName = "thinghistory"
const DefaultEventCollectionName = "events"
const DefaultActionCollectionName = "actions"
const DefaultLatestCollectionName = "latest"

// MongoHistoryStoreServer store uses MongoDB to store events, actions, and properties in time-series collections.
// This implements the client.IHistory interface
type MongoHistoryStoreServer struct {
	config config.HistoryStoreConfig

	// Client connection to the data store
	store *mongo.Client
	// database instance
	storeDB *mongo.Database

	// use a separate table for 'latest' events instead of a query on the time series
	// MongoDB query aggregate with sort and group is not performant and has memory bugs:
	// 1. https://jira.mongodb.org/browse/SERVER-4507
	// 2. https://jira.mongodb.org/browse/SERVER-68196
	// 3. https://www.mongodb.com/docs/v6.0/core/timeseries/timeseries-secondary-index/
	useSeparateLatestTable bool

	// eventCollection is the time series collection of events
	eventCollection *mongo.Collection

	// actionCollection is the time series collection of actions
	actionCollection *mongo.Collection

	// latestEvents is the collection with latest events
	latestEvents *mongo.Collection

	// startTime is the time this service started
	startTime time.Time
}

// AddAction adds a new action to the history store
func (srv *MongoHistoryStoreServer) AddAction(ctx context.Context,
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
func (srv *MongoHistoryStoreServer) AddEvent(
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
func (srv *MongoHistoryStoreServer) AddEvents(ctx context.Context,
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
func (srv *MongoHistoryStoreServer) addLatest(ctx context.Context,
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

// Delete the history database and disconnect from the store.
// Call Start to recreate it.
func (srv *MongoHistoryStoreServer) Delete() error {
	logrus.Warning("Deleting the history database")
	ctx := context.Background()

	//err := srv.store.Connect(ctx)
	//if err != nil {
	//	logrus.Error(err)
	//	return err
	//}
	time.Sleep(time.Second)
	db := srv.store.Database(srv.config.DatabaseName)
	err := db.Drop(ctx)
	if err != nil {
		logrus.Error(err)
	}
	err = srv.store.Disconnect(ctx)
	if err != nil {
		logrus.Error(err)
	}
	return err
}

// getHistory returns the request history from a collection
func (srv *MongoHistoryStoreServer) getHistory(ctx context.Context,
	collection *mongo.Collection,
	thingID string, valueName string, after string, before string, limit int) ([]thing.ThingValue, error) {

	var hist = make([]thing.ThingValue, 0)

	filter := bson.M{
		"thingID": thingID,
	}

	if after != "" {
		timeAfter, err := time.Parse(time.RFC3339, after)
		if err != nil {
			logrus.Infof("Invalid 'After' time: %s", err)
			return nil, err
		}
		timeAfterBson := primitive.NewDateTimeFromTime(timeAfter)
		filter["after"] = timeAfterBson
	}
	if before != "" {
		timeBefore, err := time.Parse(time.RFC3339, before)
		if err != nil {
			logrus.Infof("Invalid 'Before' time: %s", err)
			return nil, err
		}
		timeBeforeBson := primitive.NewDateTimeFromTime(timeBefore)
		filter["before"] = timeBeforeBson
	}
	if valueName != "" {
		filter["name"] = valueName
	}

	cursor, err := collection.Find(ctx, filter)
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

// GetActionHistory returns the action request history of a Thing
func (srv *MongoHistoryStoreServer) GetActionHistory(ctx context.Context,
	thingID string, actionName string, after string, before string, limit int) ([]thing.ThingValue, error) {

	return srv.getHistory(ctx, srv.actionCollection, thingID, actionName, after, before, limit)
}

// GetEventHistory returns the event history of a Thing
func (srv *MongoHistoryStoreServer) GetEventHistory(ctx context.Context,
	thingID string, eventName string, after string, before string, limit int) ([]thing.ThingValue, error) {

	return srv.getHistory(ctx, srv.eventCollection, thingID, eventName, after, before, limit)
}

// getLatestValuesFromTimeSeries using aggregate pipeline
// NOTE: MONGODB DOESN'T SCALE. 1 million records, 100 things, 10 sensor names
// takes a whopping 10 seconds to complete.
func (srv *MongoHistoryStoreServer) getLatestValuesFromTimeSeries(
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
func (srv *MongoHistoryStoreServer) getLatestValuesFromCollection(
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
func (srv *MongoHistoryStoreServer) GetLatestEvents(ctx context.Context,
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
func (srv *MongoHistoryStoreServer) Info(ctx context.Context) (info history.StoreInfo, err error) {
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

// setup creates missing collections in the database
func (srv *MongoHistoryStoreServer) setup(ctx context.Context) error {

	// create the database and add time series collections
	if srv.storeDB == nil {
		srv.storeDB = srv.store.Database(srv.config.DatabaseName)
	}
	// prepare options
	tso := &options.TimeSeriesOptions{
		TimeField: "timestamp",
	}
	tso.SetMetaField("metadata")

	// A granularity of hours is best if one sample per minute is received per sensor
	// choosing seconds will increase read times as many buckets need to be read.
	// choosing hours will increase write times if more samples are received as many steps are needed to add to a bucket.
	// See also this slideshare on choosing granularity:
	//   https://www.slideshare.net/mongodb/mongodb-for-time-series-data-setting-the-stage-for-sensor-management
	// tbd should a collection per sensor type name be used to match granularity?
	// setting this to hours will reduce query memory consumption
	//tso.SetGranularity("minutes") // write in minute buckets
	// for 1 sample per minute, eg 60 samples per hour, use granularity hours for read performance
	tso.SetGranularity("hours")
	co := &options.CreateCollectionOptions{}
	co.SetTimeSeriesOptions(tso)

	// events time series collection
	filter := bson.M{"name": DefaultEventCollectionName, "type": "timeseries"}
	names, err := srv.storeDB.ListCollectionNames(ctx, filter)
	if len(names) == 0 && err == nil {
		logrus.Warning("Creating the events time series")
		err = srv.storeDB.CreateCollection(ctx, DefaultEventCollectionName, co)

		// secondary index to improve sort speed using metadata.name, time
		// https://www.mongodb.com/docs/v6.0/core/timeseries/timeseries-secondary-index/
		c := srv.storeDB.Collection(DefaultEventCollectionName)
		nameIndex := mongo.IndexModel{Keys: bson.D{
			{"metadata.name", 1},
			{"timestamp", -1},
		}, Options: nil}
		indexName, err2 := c.Indexes().CreateOne(ctx, nameIndex)
		_ = indexName
		err = err2

		//speed up match on thingID
		//thingIDIndex := mongo.IndexModel{Keys: bson.D{
		//	{"thingID", 1},
		//}, Options: nil}
		//indexName, err2 = c.Indexes().CreateOne(ctx, thingIDIndex)
		//_ = indexName
		//err = err2

	}
	// actions time series collection
	filter = bson.M{"name": DefaultActionCollectionName, "type": "timeseries"}
	names, _ = srv.storeDB.ListCollectionNames(ctx, filter)
	if len(names) == 0 && err == nil {
		logrus.Warning("Creating the actions time series")
		err = srv.storeDB.CreateCollection(ctx, DefaultActionCollectionName, co)
	}

	// collection of latest thing values indexed by thingID
	if srv.useSeparateLatestTable {
		logrus.Infof("using a separate table for tracking 'latest' events")
		filter = bson.M{"name": DefaultLatestCollectionName}
		names, _ = srv.storeDB.ListCollectionNames(ctx, filter)
		if len(names) == 0 && err == nil {
			logrus.Warning("Creating the thing properties collection")
			latestOpts := &options.CreateCollectionOptions{}
			err = srv.storeDB.CreateCollection(ctx, DefaultLatestCollectionName, latestOpts)
			lc := srv.storeDB.Collection(DefaultLatestCollectionName)
			thingIDIndex := mongo.IndexModel{Keys: bson.M{"thingID": 1}, Options: nil}
			indexName, err2 := lc.Indexes().CreateOne(ctx, thingIDIndex)
			err = err2
			logrus.Infof("creating index '%s' on thing latest value collection", indexName)
		}
	} else {
		logrus.Infof("using the timeseries for getting 'latest' events")
	}
	if err != nil {
		logrus.Errorf("failed creating MongoDB time series collections: %s", err)
		return err
	}
	return err
}

// Start connect to the DB server.
// This will setup the database if the collections haven't been created yet.
// Start must be called before any other method, including Setup or Delete
func (srv *MongoHistoryStoreServer) Start() (err error) {
	logrus.Infof("Connecting to the database")
	srv.startTime = time.Now()
	srv.store, err = mongo.NewClient(options.Client().ApplyURI(srv.config.DatabaseURL))
	if err == nil {
		err = srv.store.Connect(nil)
	}
	if err != nil {
		logrus.Errorf("failed to connect to history DB on %s: %s", srv.config.DatabaseURL, err)
		return err
	}
	srv.storeDB = srv.store.Database(srv.config.DatabaseName)

	// create the collections if they don't exist
	ctx, cf := context.WithTimeout(context.Background(), time.Second*300)
	err = srv.setup(ctx)
	if err != nil {
		cf()
		return err
	}

	srv.eventCollection = srv.storeDB.Collection(DefaultEventCollectionName)
	srv.actionCollection = srv.storeDB.Collection(DefaultActionCollectionName)
	srv.latestEvents = srv.storeDB.Collection(DefaultLatestCollectionName)

	// last, populate the most recent property values
	//pipeline := `["$group": {"thingID": ]`
	//cursor, err := srv.eventCollection.Aggregate(ctx, pipeline)
	//

	cf()
	return err
}

// Stop disconnects from the DB server
// Call Start to reconnect.
func (srv *MongoHistoryStoreServer) Stop() error {
	logrus.Infof("Disconnecting from the database")
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	err := srv.store.Disconnect(ctx)
	cf()
	return err
}

// NewMongoHistoryStoreServer creates a service to access events, actions and properties in the store
// Call Start() when ready to use the store.
//  dbConfig contains the database connection settings
func NewMongoHistoryStoreServer(svcConfig config.HistoryStoreConfig) *MongoHistoryStoreServer {

	if svcConfig.DatabaseName == "" {
		svcConfig.DatabaseName = DefaultStoreName
	}

	srv := &MongoHistoryStoreServer{
		config:                 svcConfig,
		useSeparateLatestTable: true,
		startTime:              time.Now(),
	}
	return srv
}