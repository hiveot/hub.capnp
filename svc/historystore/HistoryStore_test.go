package historystore_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost.grpc/go/svc"
	"github.com/wostzone/wost.grpc/go/thing"
	"svc/historystore"
)

const storeName = "test"
const storeURL = "mongodb://localhost:27017"
const thingIDPrefix = "thing-"

var names = []string{"temperature", "humidity", "pressure", "wind", "speed", "switch", "location", "sensor-A", "sensor-B", "sensor-C"}
var testItems = make(map[string]*thing.ThingValue)
var highestName = make(map[string]*thing.ThingValue)

// add some history to the store
func addHistory(store *historystore.HistoryStoreServer, count int) {

	for i := 0; i < count; i++ {
		//randomID := rand.Intn(3)
		randomID := 1
		randomName := rand.Intn(10)
		randomValue := rand.Float64() * 100
		randomSeconds := time.Duration(rand.Intn(36000)) * time.Second
		randomTime := time.Now().Add(-randomSeconds).Format(time.RFC3339)
		ev := &thing.ThingValue{
			ThingID: thingIDPrefix + strconv.Itoa(randomID),
			Name:    names[randomName],
			Value:   fmt.Sprintf("%2.3f", randomValue),
			Created: randomTime,
			ValueID: strconv.Itoa(i),
		}
		testItems[ev.ValueID] = ev
		// track the actual most recent event for the name for thing 3
		if randomID == 1 {
			if highestName[ev.Name] == nil ||
				highestName[ev.Name].Created < ev.Created {
				highestName[ev.Name] = ev
			}
		}
		_, _ = store.AddEvent(nil, ev)
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
	store := historystore.NewHistoryStoreServer(storeURL, storeName)
	assert.NotNil(t, store)
	err := store.Start()
	if assert.NoError(t, err) {
		err = store.Stop()
		assert.NoError(t, err)
		err := store.Start()
		assert.NoError(t, err)
	}
	err = store.Delete()
	assert.NoError(t, err)
}

func TestAddGetEvent(t *testing.T) {
	const id1 = "thing1"
	const id2 = "thing2"
	const evName1 = "temperature"
	const evName2 = "humidity"
	store := historystore.NewHistoryStoreServer(storeURL, storeName)
	_ = store.Start()
	ctx := context.Background()
	// add events for thing 1
	_, err := store.AddEvent(ctx,
		&thing.ThingValue{ThingID: id1, Name: evName1, Value: "12.5"},
	)
	assert.NoError(t, err)
	_, err = store.AddEvent(ctx,
		&thing.ThingValue{ThingID: id1, Name: evName2, Value: "70"},
	)
	assert.NoError(t, err)
	// add events for thing 2
	_, err = store.AddEvent(ctx,
		&thing.ThingValue{ThingID: id2, Name: evName2, Value: "50"},
	)
	assert.NoError(t, err)
	_, err = store.AddEvent(ctx,
		&thing.ThingValue{ThingID: id2, Name: evName1, Value: "17.5"},
	)
	assert.NoError(t, err)

	// query all events of thing 1
	args := &svc.History_Args{ThingID: id1}
	res, err := store.GetEventHistory(ctx, args)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res.Values))
	assert.Equal(t, id1, res.Values[0].ThingID)
	assert.Equal(t, evName1, res.Values[0].Name)

	// query temperatures of thing 2
	args = &svc.History_Args{ThingID: id2, Name: evName1}
	res, err = store.GetEventHistory(ctx, args)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res.Values))

	_ = store.Delete()
}

func TestEventPerf(t *testing.T) {
	const id1 = "thing1"
	const evName = "event-"
	const nrRecords = 1 // 1 sec to write, 16msec to read
	store := historystore.NewHistoryStoreServer(storeURL, storeName)
	_ = store.Start()
	ctx := context.Background()

	// test adding records
	evData := `{"temperature":"12.5"}`
	t1 := time.Now()
	for i := 0; i < nrRecords; i++ {
		ev := &thing.ThingValue{
			ThingID: id1,
			Created: time.Now().Format(time.RFC3339),
			Name:    evName + strconv.Itoa(i),
			Value:   evData}

		_, err := store.AddEvent(ctx, ev)
		assert.NoError(t, err)
	}
	d1 := time.Now().Sub(t1)
	t.Logf("Adding %d events: %d msec", nrRecords, d1.Milliseconds())

	// test reading records
	t2 := time.Now()
	afterTime := time.Now().Add(-time.Minute).Format(time.RFC3339)
	args := &svc.History_Args{
		ThingID: id1,
		Before:  afterTime,
	}
	res, err := store.GetEventHistory(ctx, args)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	d2 := time.Now().Sub(t2)
	t.Logf("Reading %d events: %d msec", len(res.Values), d2.Milliseconds())

	_ = store.Delete()
}

func TestAddGetAction(t *testing.T) {
	const id1 = "thing1"
	const name = "action1"
	store := historystore.NewHistoryStoreServer(storeURL, storeName)
	_ = store.Start()
	ctx := context.Background()
	actionData := `{"switch":"on"}`
	action := &thing.ThingValue{
		ThingID:  id1,
		ActionID: "action1",
		//Created:   time.Now().Format(time.RFC3339),
		Name:  name,
		Value: actionData}

	_, err := store.AddAction(ctx, action)
	assert.NoError(t, err)
	_, err = store.AddAction(ctx, action)
	assert.NoError(t, err)

	args := &svc.History_Args{
		ThingID: id1,
	}
	res, err := store.GetActionHistory(ctx, args)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Greater(t, len(res.Values), 1)

	_ = store.Delete()
}

func TestGetLatest(t *testing.T) {
	const id1 = thingIDPrefix + "1" // matches a percentage of the random things
	const name = "action1"
	store := historystore.NewHistoryStoreServer(storeURL, storeName)
	_ = store.Start()
	addHistory(store, 10000)

	ctx := context.Background()
	args := &svc.GetLatest_Args{
		ThingID: id1,
	}
	t1 := time.Now()
	res, err := store.GetLatestValues(ctx, args)
	d1 := time.Now().Sub(t1)
	logrus.Infof("Duration: %d msec", d1.Milliseconds())
	assert.NotNil(t, res)
	if !assert.NoError(t, err) {
		_ = store.Delete()
		return
	}

	t.Logf("Received %d values", len(res.PropValues))
	assert.Greater(t, len(res.PropValues), 0)
	// compare the results with the highest value tracked during creation of the test data
	for _, val := range res.PropValues {
		logrus.Infof("Result %s: %v", val.ValueID, val)
		highest := highestName[val.Name]
		logrus.Infof("Expect %s: %v", highest.ValueID, highest)
		assert.Equal(t, highest.ValueID, val.ValueID)
	}

	_ = store.Delete()
}
