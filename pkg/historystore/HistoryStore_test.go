// This requires a local unsecured MongoDB instance
package main_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/historystore/adapter"
	"github.com/hiveot/hub/pkg/historystore/client"
	"github.com/hiveot/hub/pkg/historystore/config"
	"github.com/hiveot/hub/pkg/historystore/mongohs"
)

const thingIDPrefix = "thing-"

// when testing using the capnp RPC
const testAddress = "/tmp/histstore_test.socket"
const useCapnp = true

var svcConfig = config.HistoryStoreConfig{
	DatabaseType:    "mongodb",
	DatabaseName:    "test",
	DatabaseURL:     config.DefaultDBURL,
	LoginID:         "",
	Password:        "",
	CertificateFile: "",
}

var names = []string{"temperature", "humidity", "pressure", "wind", "speed", "switch", "location", "sensor-A", "sensor-B", "sensor-C"}
var testItems = make(map[string]thing.ThingValue)
var highestName = make(map[string]thing.ThingValue)

// Create a new store, delete if it already exists
func newStore() client.IHistoryStore {
	store := mongohs.NewMongoHistoryStoreServer(svcConfig)
	// start to delete the store
	_ = store.Start()
	_ = store.Delete()
	// start to recreate the store
	err := store.Start()
	if err != nil {
		logrus.Fatalf("Failed starting the store server: %s", err)
	}

	// optionally test with capnp RPC
	if useCapnp {
		_ = syscall.Unlink(testAddress)
		lis, _ := net.Listen("unix", testAddress)
		go adapter.StartHistoryStoreCapnpAdapter(lis, store)

		cl, err := client.NewHistoryStoreCapnpClient(testAddress, true)
		if err != nil {
			logrus.Fatalf("Failed starting capnp client: %s", err)
		}
		return cl
	}

	return store
}

//func stopStore(store client.IHistoryStore) error {
//	return store.(*mongohs.MongoHistoryStoreServer).Stop()
//}

// add some history to the store
func addHistory(store client.IHistoryStore,
	count int, nrThings int, timespanSec int) {
	var batchSize = 1000
	if batchSize > count {
		batchSize = count
	}

	// use add multiple in 100's
	for i := 0; i < count/batchSize; i++ {
		evList := make([]thing.ThingValue, 0)
		for j := 0; j < batchSize; j++ {
			randomID := rand.Intn(nrThings)
			randomName := rand.Intn(10)
			randomValue := rand.Float64() * 100
			randomSeconds := time.Duration(rand.Intn(timespanSec)) * time.Second
			randomTime := time.Now().Add(-randomSeconds).Format(time.RFC3339)
			ev := thing.ThingValue{
				ThingID:   thingIDPrefix + strconv.Itoa(randomID),
				Name:      names[randomName],
				ValueJSON: fmt.Sprintf("%2.3f", randomValue),
				Created:   randomTime,
			}
			// track the actual most recent event for the name for thing 3
			if randomID == 0 {
				if _, exists := highestName[ev.Name]; !exists ||
					highestName[ev.Name].Created < ev.Created {
					highestName[ev.Name] = ev
				}
			}
			evList = append(evList, ev)
		}
		_ = store.AddEvents(nil, evList)
	}
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

// Test creating and deleting the history database
// This requires a local unsecured MongoDB instance
func TestCreateDelete(t *testing.T) {
	store := newStore()
	if assert.NotNil(t, store) {
		//err := stopStore(store)
		//assert.NoError(t, err)
		store = newStore()
	}
	assert.NotNil(t, store)
}

func TestAddGetEvent(t *testing.T) {
	const id1 = "thing1"
	const id2 = "thing2"
	const evName1 = "temperature"
	const evName2 = "humidity"
	store := newStore()
	ctx := context.Background()
	// add events for thing 1
	err := store.AddEvent(ctx,
		thing.ThingValue{ThingID: id1, Name: evName1, ValueJSON: "12.5"},
	)
	assert.NoError(t, err)
	err = store.AddEvent(ctx,
		thing.ThingValue{ThingID: id1, Name: evName2, ValueJSON: "70"},
	)
	assert.NoError(t, err)
	// add events for thing 2
	err = store.AddEvent(ctx,
		thing.ThingValue{ThingID: id2, Name: evName2, ValueJSON: "50"},
	)
	assert.NoError(t, err)
	err = store.AddEvent(ctx,
		thing.ThingValue{ThingID: id2, Name: evName1, ValueJSON: "17.5"},
	)
	assert.NoError(t, err)

	// query all events of thing 1
	res, err := store.GetEventHistory(ctx, id1, "", "", "", 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(res))
	assert.Equal(t, id1, res[0].ThingID)
	assert.Equal(t, evName1, res[0].Name)

	// query temperatures of thing 2
	res, err = store.GetEventHistory(ctx, id2, evName1, "", "", 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	latestEvents, err := store.GetLatestEvents(ctx, id1)
	assert.NoError(t, err)
	assert.True(t, len(latestEvents) > 0)
}

func TestEventPerf(t *testing.T) {
	const id1 = "thing-1"
	const evName = "event-"
	const nrRecords = 10000 // 10000 recs: 6sec to write, 45msec to read
	store := newStore()

	//addHistory(store, 10000, 1000)

	ctx := context.Background()

	// test adding records
	evData := `{"temperature":"12.5"}`
	t1 := time.Now()
	for i := 0; i < nrRecords; i++ {
		//randomSeconds := time.Duration(rand.Intn(36000)) * time.Second
		//randomTime := time.Now().Add(-randomSeconds).Format(time.RFC3339)
		randomName := rand.Intn(10)
		ev := thing.ThingValue{
			ThingID: id1,
			Created: time.Now().Format(time.RFC3339),
			//Created: randomTime,
			Name:      names[randomName],
			ValueJSON: evData}

		err := store.AddEvent(ctx, ev)
		require.NoError(t, err)
	}
	d1 := time.Now().Sub(t1)
	t.Logf("Adding %d events: %d msec", nrRecords, d1.Milliseconds())

	// test reading records
	t2 := time.Now()
	//afterTime := time.Now().Add(-time.Hour * 600).Format(time.RFC3339)
	res, err := store.GetEventHistory(ctx, id1, "", "", "", 0)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	d2 := time.Now().Sub(t2)
	t.Logf("Reading %d events: %d msec", len(res), d2.Milliseconds())
}

func TestGetLatest(t *testing.T) {
	const id1 = thingIDPrefix + "0" // matches a percentage of the random things
	const name = "action1"
	store := newStore()

	// 10 sensors -> 1 sample per minute, 60 per hour -> 600
	addHistory(store, 1000000, 1, 3600*24*30)

	ctx := context.Background()
	t1 := time.Now()
	values, err := store.GetLatestEvents(ctx, id1)
	d1 := time.Now().Sub(t1)
	logrus.Infof("Duration: %d msec", d1.Milliseconds())
	assert.NotNil(t, values)
	if !assert.NoError(t, err) {
		return
	}

	t.Logf("Received %d values", len(values))
	assert.Greater(t, len(values), 0)
	// compare the results with the highest value tracked during creation of the test data
	for _, val := range values {
		logrus.Infof("Result %s: %v", val.Created, val)
		highest := highestName[val.Name]
		if assert.NotNil(t, highest) {
			logrus.Infof("Expect %s: %v", highest.Created, highest)
			assert.Equal(t, highest.Created, val.Created)
		}
	}
}

func TestAddGetAction(t *testing.T) {
	const id1 = "thing1"
	const name = "action1"
	store := newStore()
	ctx := context.Background()
	actionData := `{"switch":"on"}`
	action := thing.ThingValue{
		ThingID: id1,
		//Created:   time.Now().Format(time.RFC3339),
		Name:      name,
		ValueJSON: actionData}

	err := store.AddAction(ctx, action)
	assert.NoError(t, err)
	err = store.AddAction(ctx, action)
	assert.NoError(t, err)

	actions, err := store.GetActionHistory(ctx, id1, "", "", "", 0)
	assert.NoError(t, err)
	assert.NotNil(t, actions)
	assert.Greater(t, len(actions), 1)
}

func TestGetInfo(t *testing.T) {
	store := newStore()
	addHistory(store, 20, 5, 1000)
	ctx := context.Background()

	info, err := store.Info(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 20, info.NrEvents)
}
