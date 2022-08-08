package historystore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/wostzone/wost.grpc/go/svc"
	"github.com/wostzone/wost.grpc/go/thing"
)

const TimeStampField = "timestamp"
const DefaultStoreName = "thinghistory"
const DefaultEventCollectionName = "events"
const DefaultActionCollectionName = "actions"
const DefaultLatestCollectionName = "latest"

// HistoryStoreServer implements the svc.HistoryStoreServer interface
// This store uses MongoDB to store events, actions, and properties in time-series collections.
//
type HistoryStoreServer struct {
	svc.UnimplementedHistoryStoreServer
	// Client connection to the data store
	store *mongo.Client
	// database instance
	storeDB *mongo.Database
	// storeURL is the MongoDB connection URL
	storeURL string
	// storeName is the MongoDB database name of the history store
	storeName string

	// eventCollection is the time series collection of events
	eventCollection *mongo.Collection

	// actionCollection is the time series collection of actions
	actionCollection *mongo.Collection

	// latestCollection is the collection of Thing documents with latest properties
	latestCollection *mongo.Collection
}

// AddAction adds a new action to the history store
func (srv *HistoryStoreServer) AddAction(ctx context.Context, args *thing.ThingValue) (*emptypb.Empty, error) {
	// Name and ThingID are required fields
	if args.Name == "" || args.ThingID == "" {
		err := fmt.Errorf("missing name or thingID")
		logrus.Warning(err)
		return nil, err
	}
	if args.Created == "" {
		args.Created = time.Now().UTC().Format(time.RFC3339)
	}

	// It would be nice to simply use bson marshal, but that isn't possible as the
	// required timestamp needs to be added in BSON format.
	createdTime, err := time.Parse(time.RFC3339, args.Created)
	timestamp := primitive.NewDateTimeFromTime(createdTime)
	evBson := bson.M{
		TimeStampField: timestamp,
		"metadata":     bson.M{"thingID": args.ThingID, "name": args.Name},
		"name":         args.Name,
		"thingID":      args.ThingID,
		"valueID":      args.ValueID,
		"value":        args.Value,
		"created":      args.Created,
		"actionID":     args.ActionID,
	}
	res, err := srv.actionCollection.InsertOne(ctx, evBson)
	_ = res
	return nil, err
}

// AddEvent adds a new event to the history store
// The event 'created' field will be used as timestamp after parsing it using time.RFC3339
func (srv *HistoryStoreServer) AddEvent(ctx context.Context, event *thing.ThingValue) (*emptypb.Empty, error) {
	// Name and ThingID are required fields
	if event.Name == "" || event.ThingID == "" {
		err := fmt.Errorf("missing name or thingID")
		logrus.Warning(err)
		return nil, err
	}
	if event.Created == "" {
		event.Created = time.Now().UTC().Format(time.RFC3339)
	}
	if event.ValueID == "" {
		event.ValueID = uuid.New().String()
	}

	// It would be nice to simply use bson marshal, but that isn't possible as the
	// required timestamp needs to be added in BSON format.
	//createdTime, err := time.Parse("2006-01-02T15:04:05-07:00", event.Created)
	createdTime, err := time.Parse(time.RFC3339, event.Created)
	timestamp := primitive.NewDateTimeFromTime(createdTime)
	evBson := bson.M{
		TimeStampField: timestamp,
		"metadata":     bson.M{"thingID": event.ThingID},
		//"metadata":     bson.M{"thingID": event.ThingID, "name": event.Name},
		"name":     event.Name,
		"thingID":  event.ThingID,
		"valueID":  event.ValueID,
		"value":    event.Value,
		"created":  event.Created,
		"actionID": event.ActionID,
	}
	res, err := srv.eventCollection.InsertOne(ctx, evBson)
	_ = res
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	// Last, track the event value in the 'latest' collection of the Thing properties.
	// It is possible that events arrive out of order so the created date must be newer
	// than the existing date
	filter := bson.D{
		{"thingID", event.ThingID},
	}

	// translation:
	// if event.Created > {document}[event.Name].created {
	//    {document}[event.Name] = event
	// }
	pipeline := bson.A{
		bson.M{"$set": bson.M{event.Name: bson.M{"$cond": bson.A{
			bson.M{"$gt": bson.A{
				event.Created, "$" + event.Name + ".created",
			}},
			event,            // replace with new value
			"$" + event.Name, // or keep existing
		}}}},
	}
	opts := options.UpdateOptions{}
	opts.SetUpsert(true)
	res2, err := srv.latestCollection.UpdateOne(ctx, filter, pipeline, &opts)
	_ = res2
	//--- end test 2
	if err != nil {
		logrus.Error(err)
	}
	return nil, err
}

// Delete the history database and disconnect from the store.
// Call Start to recreate it.
func (srv *HistoryStoreServer) Delete() error {
	logrus.Warning("Deleting the history database")
	ctx := context.Background()
	//err := srv.store.Connect(ctx)
	//if err != nil {
	//	logrus.Error(err)
	//	return err
	//}
	time.Sleep(time.Second)
	db := srv.store.Database(srv.storeName)
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

// GetActionHistory returns the action request history of a Thing
func (srv *HistoryStoreServer) GetActionHistory(ctx context.Context, args *svc.History_Args) (*svc.ValueHistory, error) {
	var actions = make([]*thing.ThingValue, 0)

	// Is this the right way to get the data? Why can't it unmarshal directly?
	filter := bson.M{"thingID": args.ThingID}
	cursor, err := srv.actionCollection.Find(ctx, filter)
	defer cursor.Close(ctx)
	res := &svc.ValueHistory{
		Values: actions,
	}
	for cursor.Next(ctx) {
		thingAction := thing.ThingValue{}
		err = cursor.Decode(&thingAction)
		res.Values = append(res.Values, &thingAction)
	}
	return res, err
}

// GetEventHistory returns the event history of a Thing
func (srv *HistoryStoreServer) GetEventHistory(ctx context.Context, args *svc.History_Args) (*svc.ValueHistory, error) {
	var events = make([]*thing.ThingValue, 0)

	// Is this the right way to get the data?
	filter := bson.M{
		"thingID": args.ThingID,
	}

	if args.After != "" {
		timeAfter, err := time.Parse(time.RFC3339, args.After)
		if err != nil {
			logrus.Infof("Invalid 'After' time: %s", err)
			return nil, err
		}
		timeAfterBson := primitive.NewDateTimeFromTime(timeAfter)
		filter["after"] = timeAfterBson
	}
	if args.Before != "" {
		timeBefore, err := time.Parse(time.RFC3339, args.Before)
		if err != nil {
			logrus.Infof("Invalid 'Before' time: %s", err)
			return nil, err
		}
		timeBeforeBson := primitive.NewDateTimeFromTime(timeBefore)
		filter["before"] = timeBeforeBson
	}
	if args.Name != "" {
		filter["name"] = args.Name
	}
	cursor, err := srv.eventCollection.Find(ctx, filter)
	defer cursor.Close(ctx)
	res := &svc.ValueHistory{
		Values: events,
	}
	for cursor.Next(ctx) {
		thingEvent := thing.ThingValue{}
		err = cursor.Decode(&thingEvent)
		res.Values = append(res.Values, &thingEvent)
	}
	return res, err
}

// getLatestValuesFromTimeSeries using aggregate pipeline
func (srv *HistoryStoreServer) getLatestValuesFromTimeSeries(
	ctx context.Context, thingID string) (map[string]*thing.ThingValue, error) {

	values := make(map[string]*thing.ThingValue)
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
				{"timestamp", -1},
				//{"control.max.timestamp", -1},
			},
		},
	}
	// grouping doesn't take advantage of sorted sequences
	// see: https://jira.mongodb.org/browse/SERVER-4507
	groupStage := bson.D{
		{"$group",
			bson.D{
				{"_id", "$name"},
				//{"objectID", bson.M{"$first": "$_id"}},
				{"name", bson.M{"$first": "$name"}},
				{"created", bson.M{"$first": "$created"}},
				{"value", bson.M{"$first": "$value"}},
				{"valueID", bson.M{"$first": "$valueID"}},
				{"thingID", bson.M{"$first": "$thingID"}},
			},
		},
	}
	pipeline := mongo.Pipeline{matchStage, sortStage, groupStage}
	aggOptions := &options.AggregateOptions{}
	cursor, err := srv.eventCollection.Aggregate(ctx, pipeline, aggOptions)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	count := 0
	for cursor.Next(ctx) {
		var r1 map[string]interface{}
		err = cursor.Decode(&r1)

		value := thing.ThingValue{}
		// for a small number of sensor names using FindOne is faster, but for a large number, grouping is faster

		//filter1 := bson.M{"_id": r1["objectID"]}
		//one := srv.eventCollection.FindOne(ctx, filter1)
		//one.Decode(&value)
		cursor.Decode(&value)

		//err = cursor.Decode(&value)
		values[value.Name] = &value
		count++
	}
	logrus.Infof("found %d different sensors", count)
	return values, nil
}

// GetLatestValues returns the last received event/properties of a Thing
func (srv *HistoryStoreServer) GetLatestValues(ctx context.Context,
	args *svc.GetLatest_Args) (*svc.ThingValueMap, error) {

	// the hard way
	//values, err := srv.getLatestValuesFromTimeSeries(ctx, args.ThingID)

	// the easy and faster way
	propValues := map[string]*thing.ThingValue{}

	filter := bson.M{"thingID": args.ThingID}
	res := srv.latestCollection.FindOne(ctx, filter)
	var thingValues map[string]interface{}

	err := res.Decode(&thingValues)
	delete(thingValues, "_id")
	delete(thingValues, "thingID")

	asJson, err := json.Marshal(thingValues)
	err = json.Unmarshal(asJson, &propValues)

	result := &svc.ThingValueMap{
		PropValues: propValues,
	}
	return result, err
}

// setup creates missing collections in the database
func (srv *HistoryStoreServer) setup(ctx context.Context) error {

	// create the database and add time series collections
	if srv.storeDB == nil {
		srv.storeDB = srv.store.Database(srv.storeName)
	}
	// prepare options
	tso := &options.TimeSeriesOptions{
		TimeField: "timestamp",
	}
	tso.SetMetaField("metadata")
	tso.SetGranularity("minutes")
	co := &options.CreateCollectionOptions{
		DefaultIndexOptions: nil,
		MaxDocuments:        nil,
		StorageEngine:       nil,
	}
	// FOR TESTING!!! TO BE REMOVED
	co.SetExpireAfterSeconds(3600)
	co.SetTimeSeriesOptions(tso)

	// events time series collection
	filter := bson.M{"name": DefaultEventCollectionName, "type": "timeseries"}
	names, err := srv.storeDB.ListCollectionNames(ctx, filter)
	if len(names) == 0 && err == nil {
		logrus.Warning("Creating the events time series")
		err = srv.storeDB.CreateCollection(ctx, DefaultEventCollectionName, co)
	}
	// actions time series collection
	filter = bson.M{"name": DefaultActionCollectionName, "type": "timeseries"}
	names, _ = srv.storeDB.ListCollectionNames(ctx, filter)
	if len(names) == 0 && err == nil {
		logrus.Warning("Creating the actions time series")
		err = srv.storeDB.CreateCollection(ctx, DefaultActionCollectionName, co)
	}
	// collection of latest thing values indexed by thingID
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

	if err != nil {
		logrus.Errorf("failed creating MongoDB time series collections: %s", err)
		return err
	}
	return err
}

// Start connect to the DB server.
// This will setup the database if the collections haven't been created yet.
// Start must be called before any other method, including Setup or Delete
func (srv *HistoryStoreServer) Start() error {
	logrus.Infof("Connecting to the database")
	store, err := mongo.NewClient(options.Client().ApplyURI(srv.storeURL))
	if err != nil {
		logrus.Errorf("Failed to create DB client on %s: %s", srv.storeURL, err)
		return err
	}
	srv.store = store

	err = srv.store.Connect(nil)
	if err != nil {
		logrus.Errorf("failed to connect to history DB on %s: %s", srv.storeURL, err)
		return err
	}
	srv.storeDB = srv.store.Database(srv.storeName)

	// create the collections if they don't exist
	ctx, cf := context.WithTimeout(context.Background(), time.Second*300)
	err = srv.setup(ctx)
	if err != nil {
		cf()
		return err
	}

	srv.eventCollection = srv.storeDB.Collection(DefaultEventCollectionName,
		&options.CollectionOptions{
			ReadConcern: &readconcern.ReadConcern{},
		})
	srv.actionCollection = srv.storeDB.Collection(DefaultActionCollectionName,
		&options.CollectionOptions{
			ReadConcern: &readconcern.ReadConcern{},
		})
	srv.latestCollection = srv.storeDB.Collection(DefaultLatestCollectionName,
		&options.CollectionOptions{
			ReadConcern: &readconcern.ReadConcern{},
		})

	// last, populate the most recent property values
	//pipeline := `["$group": {"thingID": ]`
	//cursor, err := srv.eventCollection.Aggregate(ctx, pipeline)
	//

	cf()
	return err
}

// Stop disconnects from the DB server
// Call Start to reconnect.
func (srv *HistoryStoreServer) Stop() error {
	logrus.Infof("Disconnecting from the database")
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	err := srv.store.Disconnect(ctx)
	cf()
	return err
}

// NewHistoryStoreServer creates a service to access events, actions and properties in the store
// Call Start() when ready to use the store.
//  storeURL is the full URL to the database
//  storeName is the database name, use "" for DefaultStoreName or "test" for testing
func NewHistoryStoreServer(storeURL string, storeName string) *HistoryStoreServer {

	if storeName == "" {
		storeName = DefaultStoreName
	}

	srv := &HistoryStoreServer{
		storeURL:  storeURL,
		storeName: storeName,
	}
	return srv
}
