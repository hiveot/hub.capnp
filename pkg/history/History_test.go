// This requires a local unsecured MongoDB instance
package history_test

import (
	"context"
	"encoding/json"
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

	"github.com/hiveot/hub.go/pkg/vocab"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/bucketstore/cmd"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpclient"
	"github.com/hiveot/hub/pkg/history/capnpserver"
	"github.com/hiveot/hub/pkg/history/config"
	"github.com/hiveot/hub/pkg/history/service"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub.go/pkg/thing"
)

const thingIDPrefix = "thing-"

// when testing using the capnp RPC
const testAddress = "/tmp/histstore_test.socket"
const testStoreDirectory = "/tmp/test-history"
const testClientID = "testclient"
const useTestCapnp = false
const HistoryStoreBackend = bucketstore.BackendKVBTree

//var svcConfig = config.HistoryStoreConfig{
//	DatabaseType:    "mongodb",
//	DatabaseName:    "test",
//	DatabaseURL:     config.DefaultDBURL,
//	LoginID:         "",
//	Password:        "",
//	ClientCertificate: "",
//}

var names = []string{"temperature", "humidity", "pressure", "wind", "speed", "switch", "location", "sensor-A", "sensor-B", "sensor-C"}

// var testItems = make(map[string]*thing.ThingValue)
//var testHighestName = make(map[string]*thing.ThingValue)

// Create a new store, delete if it already exists
func newHistoryService(useCapnp bool) (history.IHistoryService, func() error) {
	svcConfig := config.NewHistoryConfig()
	svcConfig.DatabaseName = "test"

	// create a new empty store to use
	_ = os.RemoveAll(testStoreDirectory)
	store := cmd.NewBucketStore(testStoreDirectory, testClientID, HistoryStoreBackend)
	err := store.Open()
	if err != nil {
		logrus.Fatalf("Unable to open test store: %s", err)
	}
	// start the service
	svc := service.NewHistoryService(store)
	ctx, cancelFn := context.WithCancel(context.Background())
	err = svc.Start(ctx)
	if err != nil {
		logrus.Fatalf("Failed starting the state service: %s", err)
	}

	// optionally test with capnp RPC
	if useCapnp {
		_ = syscall.Unlink(testAddress)
		srvListener, _ := net.Listen("unix", testAddress)
		go capnpserver.StartHistoryServiceCapnpServer(ctx, srvListener, svc)
		// connect the client to the server above
		clConn, _ := net.Dial("unix", testAddress)
		cl, err := capnpclient.NewHistoryCapnpClient(ctx, clConn)
		if err != nil {
			logrus.Fatalf("Failed starting capnp client: %s", err)
		}
		return cl, func() error {
			cancelFn()
			_ = svc.Stop(ctx)
			err = store.Close()
			return err
		}
	}

	return svc, func() error {
		_ = svc.Stop(ctx)
		err = store.Close()
		cancelFn()
		return err
	}
}

//func stopStore(store client.IHistory) error {
//	return store.(*mongohs.MongoHistoryServer).Stop()
//}

// generate a random batch of values for testing
func makeValueBatch(nrValues, nrThings, timespanSec int) (batch []*thing.ThingValue, highest map[string]*thing.ThingValue) {
	highest = make(map[string]*thing.ThingValue)
	valueBatch := make([]*thing.ThingValue, 0, nrValues)
	for j := 0; j < nrValues; j++ {
		randomID := rand.Intn(nrThings)
		randomName := rand.Intn(10)
		randomValue := rand.Float64() * 100
		randomSeconds := time.Duration(rand.Intn(timespanSec)) * time.Second
		randomTime := time.Now().Add(-randomSeconds).Format(vocab.ISO8601Format)
		ev := &thing.ThingValue{
			ThingID:   thingIDPrefix + strconv.Itoa(randomID),
			Name:      names[randomName],
			ValueJSON: []byte(fmt.Sprintf("%2.3f", randomValue)),
			Created:   randomTime,
		}
		// track the actual most recent event for the name for thing 3
		if randomID == 0 {
			if _, exists := highest[ev.Name]; !exists ||
				highest[ev.Name].Created < ev.Created {
				highest[ev.Name] = ev
			}
		}
		valueBatch = append(valueBatch, ev)
	}
	return valueBatch, highest
}

// add some history to the store
func addHistory(svc history.IHistoryService, count int, nrThings int, timespanSec int) (
	highest map[string]*thing.ThingValue) {

	var batchSize = 1000
	if batchSize > count {
		batchSize = count
	}
	ctx := context.Background()

	evBatch, highest := makeValueBatch(count, nrThings, timespanSec)

	// use add multiple in 100's
	for i := 0; i < count/batchSize; i++ {
		// no thingID constraint allows adding events from any thing
		capAdd := svc.CapAddAnyThing(ctx)
		start := batchSize * i
		end := batchSize * (i + 1)
		err := capAdd.AddEvents(ctx, evBatch[start:end])
		if err != nil {
			logrus.Fatalf("Problem adding events: %s", err)
		}
		capAdd.Release()
	}
	return highest
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

// Test creating and deleting the history database
// This requires a local unsecured MongoDB instance
func TestStartStop(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewHistoryConfig()
	cfg.DatabaseName = "test"
	//store := NewBucketStore()
	store := cmd.NewBucketStore(testStoreDirectory, testClientID, bucketstore.BackendKVBTree)
	err := store.Open()
	assert.NoError(t, err)
	svc := service.NewHistoryService(store)

	err = svc.Start(ctx)
	assert.NoError(t, err)

	err = svc.Stop(ctx)
	assert.NoError(t, err)

	err = store.Close()
	assert.NoError(t, err)
}

func TestAddGetEvent(t *testing.T) {
	const id1 = "thing1"
	const id2 = "thing2"
	const evTemperature = "temperature"
	const evHumidity = "humidity"
	var timeafter = ""

	store, cancelFn := newHistoryService(useTestCapnp)
	ctx := context.Background()
	fivemago := time.Now().Add(-time.Minute * 5)
	fiftyfivemago := time.Now().Add(-time.Minute * 55)
	addHistory(store, 20, 3, 3600)

	// add events for thing 1
	addHistory1 := store.CapAddHistory(ctx, id1)
	readHistory1 := store.CapReadHistory(ctx, id1)

	// release and cancel is order dependent
	defer addHistory1.Release()
	defer readHistory1.Release()
	defer cancelFn()

	// add thing1 temperature from 5 minutes ago
	ev1_1 := &thing.ThingValue{ThingID: id1, Name: evTemperature,
		ValueJSON: []byte("12.5"), Created: fivemago.Format(vocab.ISO8601Format)}
	err := addHistory1.AddEvent(ctx, ev1_1)
	assert.NoError(t, err)
	// add thing1 humidity from 55 minutes ago
	ev1_2 := &thing.ThingValue{ThingID: id1, Name: evHumidity,
		ValueJSON: []byte("70"), Created: fiftyfivemago.Format(vocab.ISO8601Format)}
	err = addHistory1.AddEvent(ctx, ev1_2)
	assert.NoError(t, err)

	// add events for thing 2, temperature and humidity
	addHistory2 := store.CapAddHistory(ctx, id2)
	// add thing2 humidity from 5 minutes ago
	ev2_1 := &thing.ThingValue{ThingID: id2, Name: evHumidity,
		ValueJSON: []byte("50"), Created: fivemago.Format(vocab.ISO8601Format)}
	err = addHistory2.AddEvent(ctx, ev2_1)
	assert.NoError(t, err)

	// add thing2 temperature from 55 minutes ago
	ev2_2 := &thing.ThingValue{ThingID: id2, Name: evTemperature,
		ValueJSON: []byte("17.5"), Created: fiftyfivemago.Format(vocab.ISO8601Format)}
	err = addHistory2.AddEvent(ctx, ev2_2)
	assert.NoError(t, err)
	addHistory2.Release()

	// Test 1: get events of thing 1 older than 300 minutes ago - expect 1 humidity from 55 minutes ago
	cursor1 := readHistory1.GetEventHistory(ctx, "")
	// seek must return the thing humidity added 55 minutes ago, not 5 minutes ago
	timeafter = time.Now().Add(-time.Minute * 300).Format(vocab.ISO8601Format)
	res1, valid := cursor1.Seek(timeafter)
	if assert.True(t, valid) {
		assert.Equal(t, id1, res1.ThingID)
		assert.Equal(t, evHumidity, res1.Name)
		// next finds the temperature from 5 minutes ago
		res1, valid = cursor1.Next()
		assert.True(t, valid)
		assert.Equal(t, evTemperature, res1.Name)
	}
	// Test 2: get events of thing 1 newer than 30 minutes ago - expect 1 temperature
	timeafter = time.Now().Add(-time.Minute * 30).Format(vocab.ISO8601Format)

	// do we need to get a new cursor?
	//readHistory = store.CapReadHistory()
	res2, valid := cursor1.Seek(timeafter)
	if assert.True(t, valid) {
		assert.Equal(t, id1, res2.ThingID)        // must match the filtered id1
		assert.Equal(t, evTemperature, res2.Name) // must match evTemperature from 5 minutes ago
		assert.Equal(t, fivemago.Format(vocab.ISO8601Format), res2.Created)
	}
	cursor1.Release()
	cursor1 = nil

	// Test 3: get first temperature of thing 2 - expect 1 result
	readHistory2 := store.CapReadHistory(ctx, id2)
	cursor2 := readHistory2.GetEventHistory(ctx, "")
	res3, valid := cursor2.First()
	cursor2.Release()
	cursor2 = nil
	require.True(t, valid)
	assert.Equal(t, evTemperature, res3.Name)
}

func TestAddPropertiesEvent(t *testing.T) {
	const count = 1000
	const id1 = thingIDPrefix + "0" // matches a percentage of the random things
	const temp1 = "55"
	store, closeFn := newHistoryService(useTestCapnp)

	ctx := context.Background()
	addHist := store.CapAddHistory(ctx, id1)
	readHist := store.CapReadHistory(ctx, id1)

	action1 := &thing.ThingValue{
		ThingID:   id1,
		Name:      vocab.PropNameSwitch,
		ValueJSON: []byte("on"),
	}
	event1 := &thing.ThingValue{
		ThingID:   id1,
		Name:      vocab.PropNameTemperature,
		ValueJSON: []byte(temp1),
	}
	badEvent1 := &thing.ThingValue{
		ThingID: id1,
		Name:    "", // missing name
	}
	badEvent2 := &thing.ThingValue{
		ThingID: "fake", // wrong ID
		Name:    "name",
	}
	badEvent3 := &thing.ThingValue{
		ThingID: id1, // wrong ID
		Name:    "baddate",
		Created: "notadate",
	}
	badEvent4 := &thing.ThingValue{
		ThingID: "", // missing ID
		Name:    "temperature",
	}
	propsList := make(map[string][]byte)
	propsList[vocab.PropNameBattery] = []byte("50")
	propsList[vocab.PropNameCPULevel] = []byte("30")
	propsList[vocab.PropNameSwitch] = []byte("off")
	propsValue, _ := json.Marshal(propsList)
	props1 := &thing.ThingValue{
		ThingID:   id1,
		Name:      history.EventNameProperties,
		ValueJSON: propsValue,
	}

	// in total add 5 properties
	err := addHist.AddAction(ctx, action1)
	assert.NoError(t, err)
	err = addHist.AddEvent(ctx, event1)
	assert.NoError(t, err)
	err = addHist.AddEvent(ctx, props1) // props has 3 values
	assert.NoError(t, err)

	// and some bad values
	err = addHist.AddEvent(ctx, badEvent1)
	assert.Error(t, err)
	err = addHist.AddEvent(ctx, badEvent2)
	assert.Error(t, err)
	err = addHist.AddEvent(ctx, badEvent3) // bad date is recovered
	assert.NoError(t, err)
	err = addHist.AddEvent(ctx, badEvent4)
	assert.Error(t, err)
	err = addHist.AddEvent(ctx, nil)
	assert.Error(t, err)
	err = addHist.AddEvents(ctx, []*thing.ThingValue{badEvent1, badEvent2, badEvent3, badEvent4})
	assert.Error(t, err)
	err = addHist.AddAction(ctx, badEvent1)
	assert.Error(t, err)
	err = addHist.AddAction(ctx, nil)
	assert.Error(t, err)

	// verify named properties from different sources
	props := readHist.GetProperties(ctx, []string{vocab.PropNameTemperature, vocab.PropNameSwitch})
	assert.Equal(t, 2, len(props))
	assert.Equal(t, props[0].Name, vocab.PropNameTemperature)
	assert.Equal(t, props[0].ValueJSON, []byte(temp1))
	assert.Equal(t, props[1].Name, vocab.PropNameSwitch)

	// restart
	readHist.Release()
	addHist.Release()
	err = closeFn()
	assert.NoError(t, err)
	backend := cmd.NewBucketStore(testStoreDirectory, testClientID, HistoryStoreBackend)
	err = backend.Open()
	assert.NoError(t, err)
	svc := service.NewHistoryService(backend)
	err = svc.Start(ctx)
	assert.NoError(t, err)

	// after closing and reopen the store the properties should still be there
	readHist = svc.CapReadHistory(ctx, id1)
	props = readHist.GetProperties(ctx, []string{vocab.PropNameTemperature, vocab.PropNameSwitch})
	assert.Equal(t, 2, len(props))
	assert.Equal(t, props[0].Name, vocab.PropNameTemperature)
	assert.Equal(t, props[0].ValueJSON, []byte(temp1))
	assert.Equal(t, props[1].Name, vocab.PropNameSwitch)

	err = svc.Stop(ctx)
	assert.NoError(t, err)
	err = backend.Close()
	assert.NoError(t, err)

}

func TestGetLatest(t *testing.T) {
	const count = 1000
	const id1 = thingIDPrefix + "0" // matches a percentage of the random things
	store, closeFn := newHistoryService(useTestCapnp)
	defer closeFn()

	ctx := context.Background()

	// 10 sensors -> 1 sample per minute, 60 per hour -> 600
	// TODO: use different timezones
	highestFromAdded := addHistory(store, count, 1, 3600*24*30)

	readHistory := store.CapReadHistory(ctx, id1)
	values := readHistory.GetProperties(ctx, nil)
	cursor := readHistory.GetEventHistory(ctx, "")
	readHistory.Release()
	readHistory = nil

	assert.NotNil(t, values)

	t.Logf("Received %d values", len(values))
	assert.Greater(t, len(values), 0, "Expected multiple properties, got none")
	// compare the results with the highest value tracked during creation of the test data
	for _, val := range values {
		logrus.Infof("Result %s: %s", val.Name, val.Created)
		highest := highestFromAdded[val.Name]
		if assert.NotNil(t, highest) {
			logrus.Infof("Expect %s: %v", highest.Name, highest.Created)
			assert.Equal(t, highest.Created, val.Created)
		}
	}
	// getting the Last should get the same result
	lastItem, valid := cursor.Last()
	highest := highestFromAdded[lastItem.Name]

	assert.True(t, valid)
	assert.Equal(t, lastItem.Created, highest.Created)
	cursor.Release()
}

func TestPrevNext(t *testing.T) {
	const count = 1000
	const id0 = thingIDPrefix + "0" // matches a percentage of the random things
	store, closeFn := newHistoryService(useTestCapnp)
	defer closeFn()

	ctx := context.Background()

	// 10 sensors -> 1 sample per minute, 60 per hour -> 600
	// TODO: use different timezones
	_ = addHistory(store, count, 1, 3600*24*30)

	readHistory := store.CapReadHistory(ctx, id0)
	cursor := readHistory.GetEventHistory(ctx, "")
	readHistory.Release()
	readHistory = nil
	assert.NotNil(t, cursor)

	// go forward
	item0, valid := cursor.First()
	assert.True(t, valid)
	assert.NotEmpty(t, item0)
	item1, valid := cursor.Next()
	assert.True(t, valid)
	assert.NotEmpty(t, item1)
	items2to11, itemsRemaining := cursor.NextN(10)
	assert.True(t, itemsRemaining)
	assert.Equal(t, 10, len(items2to11))

	// go backwards
	item10to1, itemsRemaining := cursor.PrevN(10)
	assert.True(t, valid)
	assert.Equal(t, 10, len(item10to1))

	// reached first item
	item0b, valid := cursor.Prev()
	assert.True(t, valid)
	assert.Equal(t, item0.Created, item0b.Created)

	// can't skip before the beginning of time
	iteminv, valid := cursor.Prev()
	_ = iteminv
	assert.False(t, valid)

	// seek to item11 should succeed
	item11 := items2to11[9]
	item11b, valid := cursor.Seek(item11.Created)
	assert.True(t, valid)
	assert.Equal(t, item11.Name, item11b.Name)

	cursor.Release()
}

// filter on property name
func TestPrevNextFiltered(t *testing.T) {
	const count = 1000
	const id0 = thingIDPrefix + "0" // matches a percentage of the random things
	store, closeFn := newHistoryService(useTestCapnp)
	defer closeFn()

	ctx := context.Background()

	// 10 sensors -> 1 sample per minute, 60 per hour -> 600
	// TODO: use different timezones
	_ = addHistory(store, count, 1, 3600*24*30)
	propName := names[2] // names used to generate the history

	readHistory := store.CapReadHistory(ctx, id0)
	values := readHistory.GetProperties(ctx, []string{propName})
	cursor := readHistory.GetEventHistory(ctx, propName)
	readHistory.Release()
	readHistory = nil

	assert.NotNil(t, values)
	assert.NotNil(t, cursor)

	// go forward
	item0, valid := cursor.First()
	assert.True(t, valid)
	assert.Equal(t, propName, item0.Name)
	item1, valid := cursor.Next()
	assert.True(t, valid)
	assert.Equal(t, propName, item1.Name)
	items2to11, itemsRemaining := cursor.NextN(10)
	assert.True(t, itemsRemaining)
	assert.Equal(t, 10, len(items2to11))
	assert.Equal(t, propName, items2to11[9].Name)

	// go backwards
	item10to1, itemsRemaining := cursor.PrevN(10)
	assert.True(t, valid)
	assert.Equal(t, 10, len(item10to1))

	// reached first item
	item0b, valid := cursor.Prev()
	assert.True(t, valid)
	assert.Equal(t, item0.Created, item0b.Created)
	assert.Equal(t, propName, item0b.Name)

	// can't skip before the beginning of time
	iteminv, valid := cursor.Prev()
	_ = iteminv
	assert.False(t, valid)

	// seek to item11 should succeed
	item11 := items2to11[9]
	item11b, valid := cursor.Seek(item11.Created)
	assert.True(t, valid)
	assert.Equal(t, item11.Name, item11b.Name)

	// last item should be of the name
	lastItem, valid := cursor.Last()
	assert.True(t, valid)
	assert.Equal(t, propName, lastItem.Name)

	cursor.Release()
}

func TestGetInfo(t *testing.T) {
	store, closeFn := newHistoryService(useTestCapnp)
	defer closeFn()
	addHistory(store, 1000, 5, 1000)
	ctx := context.Background()

	//info := store.Info(ctx)
	//t.Logf("Store ID:%s, records:%d", info.Id, info.NrRecords)

	readHistory := store.CapReadHistory(ctx, thingIDPrefix+"0")
	defer readHistory.Release()

	info := readHistory.Info(ctx)
	assert.NotEmpty(t, info.Engine)
	assert.NotEmpty(t, info.Id)
	t.Logf("ID:%s records:%d", info.Id, info.NrRecords)
}
