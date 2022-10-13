// Package mongohs with MongoDB based history store
// This implements the HistoryStore.proto API
package mongohs

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
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

// MongoHistoryServer store uses MongoDB to store events, actions, and properties in time-series collections.
// This implements the client.IHistory interface
type MongoHistoryServer struct {
	config config.HistoryConfig

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

// CapReadHistory provides the capability to read history
func (srv *MongoHistoryServer) CapReadHistory() history.IReadHistory {
	return srv
}

// CapUpdateHistory provides the capability to update history
func (srv *MongoHistoryServer) CapUpdateHistory() history.IUpdateHistory {
	return srv
}

// Delete the history database and disconnect from the store.
// Call Start to recreate it.
func (srv *MongoHistoryServer) Delete() error {
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
	_ = srv.store.Disconnect(ctx)
	srv.store = nil
	return err
}

// GetActionHistory returns the action request history of a Thing
func (srv *MongoHistoryServer) GetActionHistory(ctx context.Context,
	thingID string, actionName string, after string, before string, limit int) ([]thing.ThingValue, error) {

	return srv.getHistory(ctx, srv.actionCollection, thingID, actionName, after, before, limit)
}

// GetEventHistory returns the event history of a Thing
func (srv *MongoHistoryServer) GetEventHistory(ctx context.Context,
	thingID string, eventName string, after string, before string, limit int) ([]thing.ThingValue, error) {

	return srv.getHistory(ctx, srv.eventCollection, thingID, eventName, after, before, limit)
}

// setup creates missing collections in the database
// srv.storeDB must exist and be usable.
func (srv *MongoHistoryServer) setup(ctx context.Context) error {

	// create the time series collections
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
func (srv *MongoHistoryServer) Start() (err error) {
	logrus.Infof("Connecting to the mongodb database on '%s'", srv.config.DatabaseURL)
	if srv.store != nil {
		return fmt.Errorf("Store already started")
	}
	srv.startTime = time.Now()
	srv.store, err = mongo.NewClient(options.Client().ApplyURI(srv.config.DatabaseURL))
	if err == nil {
		err = srv.store.Connect(nil)
	}
	ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
	if err == nil {
		err = srv.store.Ping(ctx, nil)
	}
	if err != nil {
		logrus.Errorf("failed to connect to history DB on %s: %s", srv.config.DatabaseURL, err)
		cf()
		return err
	}
	srv.storeDB = srv.store.Database(srv.config.DatabaseName)

	// create the collections if they don't exist
	//ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
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
func (srv *MongoHistoryServer) Stop() error {
	logrus.Infof("Disconnecting from the database")
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	err := srv.store.Disconnect(ctx)
	srv.store = nil
	cf()
	return err
}

// NewMongoHistoryServer creates a service to access events, actions and properties in the store
// Call Start() when ready to use the store.
//  dbConfig contains the database connection settings
func NewMongoHistoryServer(svcConfig config.HistoryConfig) *MongoHistoryServer {
	if svcConfig.DatabaseName == "" {
		svcConfig.DatabaseName = DefaultStoreName
	}

	srv := &MongoHistoryServer{
		config:                 svcConfig,
		useSeparateLatestTable: true,
		startTime:              time.Now(),
	}
	return srv
}
