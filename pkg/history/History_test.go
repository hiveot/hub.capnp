package history_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.capnp/go/vocab"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/bucketstore/cmd"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/capnpclient"
	"github.com/hiveot/hub/pkg/history/capnpserver"
	"github.com/hiveot/hub/pkg/history/config"
	"github.com/hiveot/hub/pkg/history/service"
	"github.com/hiveot/hub/pkg/pubsub"
	service2 "github.com/hiveot/hub/pkg/pubsub/service"

	"github.com/hiveot/hub/lib/logging"

	"github.com/hiveot/hub/lib/thing"
)

const thingIDPrefix = "thing-"

// when testing using the capnp RPC
var testFolder = path.Join(os.TempDir(), "test-history")
var testSocket = path.Join(testFolder, history.ServiceName+".socket")

const testClientID = "testclient"
const useTestCapnp = true
const HistoryStoreBackend = bucketstore.BackendPebble

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
func newHistoryService(useCapnp bool) (history.IHistoryService, func()) {
	var sub pubsub.IServicePubSub
	svcConfig := config.NewHistoryConfig(testFolder)

	// create a new empty store to use
	_ = os.RemoveAll(svcConfig.Directory)
	store := cmd.NewBucketStore(testFolder, testClientID, HistoryStoreBackend)
	// start the service ; not using pubsub
	svc := service.NewHistoryService(&svcConfig, store, sub)
	ctx, cancelFn := context.WithCancel(context.Background())
	err := svc.Start()
	if err != nil {
		logrus.Fatalf("Failed starting the state service: %s", err)
	}

	// optionally test with capnp RPC
	if useCapnp {
		_ = syscall.Unlink(testSocket)
		srvListener, _ := net.Listen("unix", testSocket)
		go capnpserver.StartHistoryServiceCapnpServer(svc, srvListener)

		// connect the client to the server above
		clConn, _ := net.Dial("unix", testSocket)
		cl := capnpclient.NewHistoryCapnpClient(ctx, clConn)

		return cl, func() {
			cl.Release()
			cancelFn()
			_ = srvListener.Close()
			_ = svc.Stop()
			// give it some time to shut down before the next test
			time.Sleep(time.Millisecond)
		}
	}

	return svc, func() {
		_ = svc.Stop()
		cancelFn()
	}
}

//func stopStore(store client.IHistory) error {
//	return store.(*mongohs.MongoHistoryServer).Stop()
//}

// generate a random batch of values for testing
func makeValueBatch(publisherID string, nrValues, nrThings, timespanSec int) (batch []*thing.ThingValue, highest map[string]*thing.ThingValue) {
	highest = make(map[string]*thing.ThingValue)
	valueBatch := make([]*thing.ThingValue, 0, nrValues)
	for j := 0; j < nrValues; j++ {
		randomID := rand.Intn(nrThings)
		randomName := rand.Intn(10)
		randomValue := rand.Float64() * 100
		randomSeconds := time.Duration(rand.Intn(timespanSec)) * time.Second
		randomTime := time.Now().Add(-randomSeconds).Format(vocab.ISO8601Format)
		thingID := thingIDPrefix + strconv.Itoa(randomID)

		ev := &thing.ThingValue{
			PublisherID: publisherID,
			ThingID:     thingID,
			Name:        names[randomName],
			ValueJSON:   []byte(fmt.Sprintf("%2.3f", randomValue)),
			Created:     randomTime,
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

// add some history to the store using publisher 'device1'
func addHistory(svc history.IHistoryService, count int, nrThings int, timespanSec int) (
	highest map[string]*thing.ThingValue) {
	const publisherID = "device1"
	var batchSize = 1000
	if batchSize > count {
		batchSize = count
	}
	ctx := context.Background()

	evBatch, highest := makeValueBatch(publisherID, count, nrThings, timespanSec)

	// use add multiple in 100's
	for i := 0; i < count/batchSize; i++ {
		// no thingID constraint allows adding events from any thing
		capAdd, err := svc.CapAddHistory(ctx, "test", true)
		if err != nil {
			panic("failed cap add")
		}
		start := batchSize * i
		end := batchSize * (i + 1)
		err = capAdd.AddEvents(ctx, evBatch[start:end])
		if err != nil {
			logrus.Fatalf("Problem adding events: %s", err)
		}
		capAdd.Release()
	}
	return highest
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	_ = os.RemoveAll(testFolder)
	_ = os.MkdirAll(testFolder, 0700)

	res := m.Run()
	os.Exit(res)
}

// Test creating and deleting the history database
// This requires a local unsecured MongoDB instance
func TestStartStop(t *testing.T) {
	logrus.Info("--- TestStartStop ---")
	cfg := config.NewHistoryConfig(testFolder)

	store := cmd.NewBucketStore(cfg.Directory, testClientID, bucketstore.BackendKVBTree)
	svc := service.NewHistoryService(&cfg, store, nil)

	err := svc.Start()
	assert.NoError(t, err)

	err = svc.Stop()
	assert.NoError(t, err)
}

func TestAddGetEvent(t *testing.T) {
	logrus.Info("--- TestAddGetEvent ---")
	const device1 = "device1"
	const id1 = "thing1"
	const id2 = "thing2"
	const publisherID = "device1"
	const thing1ID = id1
	const thing2ID = id2
	const evTemperature = "temperature"
	const evHumidity = "humidity"
	var timeafter = ""

	svc, cancelFn := newHistoryService(useTestCapnp)
	ctx := context.Background()
	fivemago := time.Now().Add(-time.Minute * 5)
	fiftyfivemago := time.Now().Add(-time.Minute * 55)
	addHistory(svc, 20, 3, 3600)

	// add events for thing 1
	addHistory1, _ := svc.CapAddHistory(ctx, device1, true)
	readHistory1, _ := svc.CapReadHistory(ctx, device1, publisherID, thing1ID)

	// add thing1 temperature from 5 minutes ago
	ev1_1 := &thing.ThingValue{PublisherID: publisherID, ThingID: thing1ID, Name: evTemperature,
		ValueJSON: []byte("12.5"), Created: fivemago.Format(vocab.ISO8601Format)}
	err := addHistory1.AddEvent(ctx, ev1_1)
	assert.NoError(t, err)
	// add thing1 humidity from 55 minutes ago
	ev1_2 := &thing.ThingValue{PublisherID: publisherID, ThingID: thing1ID, Name: evHumidity,
		ValueJSON: []byte("70"), Created: fiftyfivemago.Format(vocab.ISO8601Format)}
	err = addHistory1.AddEvent(ctx, ev1_2)
	assert.NoError(t, err)

	// add events for thing 2, temperature and humidity
	addHistory2, _ := svc.CapAddHistory(ctx, device1, true)
	// add thing2 humidity from 5 minutes ago
	ev2_1 := &thing.ThingValue{PublisherID: publisherID, ThingID: thing2ID, Name: evHumidity,
		ValueJSON: []byte("50"), Created: fivemago.Format(vocab.ISO8601Format)}
	err = addHistory2.AddEvent(ctx, ev2_1)
	assert.NoError(t, err)

	// add thing2 temperature from 55 minutes ago
	ev2_2 := &thing.ThingValue{PublisherID: publisherID, ThingID: thing2ID, Name: evTemperature,
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
		assert.Equal(t, thing1ID, res1.ThingID)
		assert.Equal(t, evHumidity, res1.Name)
		// next finds the temperature from 5 minutes ago
		res1, valid = cursor1.Next()
		assert.True(t, valid)
		assert.Equal(t, evTemperature, res1.Name)
	}
	// Test 2: get events of thing 1 newer than 30 minutes ago - expect 1 temperature
	timeafter = time.Now().Add(-time.Minute * 30).Format(vocab.ISO8601Format)

	// do we need to get a new cursor?
	//readHistory = svc.CapReadHistory()
	res2, valid := cursor1.Seek(timeafter)
	if assert.True(t, valid) {
		assert.Equal(t, thing1ID, res2.ThingID)   // must match the filtered id1
		assert.Equal(t, evTemperature, res2.Name) // must match evTemperature from 5 minutes ago
		assert.Equal(t, fivemago.Format(vocab.ISO8601Format), res2.Created)
	}
	cursor1.Release()
	readHistory1.Release()
	addHistory1.Release()
	cancelFn()

	// after closing and reopening the svc the event should still be there
	store2 := cmd.NewBucketStore(testFolder, testClientID, HistoryStoreBackend)
	svcConfig2 := config.NewHistoryConfig(testFolder)
	svc2 := service.NewHistoryService(&svcConfig2, store2, nil)
	err = svc2.Start()
	require.NoError(t, err)

	// Test 3: get first temperature of thing 2 - expect 1 result
	readHistory2, _ := svc2.CapReadHistory(ctx, device1, publisherID, thing2ID)
	cursor2 := readHistory2.GetEventHistory(ctx, "")
	res3, valid := cursor2.First()
	require.True(t, valid)
	assert.Equal(t, evTemperature, res3.Name)

	cursor2.Release()
	readHistory2.Release()
	store2.Close()
}

func TestAddPropertiesEvent(t *testing.T) {
	logrus.Info("--- TestAddPropertiesEvent ---")
	const clientID = "device0"
	const thing1ID = thingIDPrefix + "0" // matches a percentage of the random things
	const publisherID = "device1"
	const temp1 = "55"
	store, closeFn := newHistoryService(useTestCapnp)

	ctx := context.Background()
	addHist, _ := store.CapAddHistory(ctx, clientID, true)
	readHist, _ := store.CapReadHistory(ctx, clientID, publisherID, thing1ID)

	action1 := &thing.ThingValue{
		PublisherID: publisherID,
		ThingID:     thing1ID,
		Name:        vocab.PropNameSwitch,
		ValueJSON:   []byte("on"),
	}
	event1 := &thing.ThingValue{
		PublisherID: publisherID,
		ThingID:     thing1ID,
		Name:        vocab.PropNameTemperature,
		ValueJSON:   []byte(temp1),
	}
	badEvent1 := &thing.ThingValue{
		PublisherID: publisherID,
		ThingID:     thing1ID,
		Name:        "", // missing name
	}
	badEvent2 := &thing.ThingValue{
		PublisherID: "", // missing publisher
		ThingID:     thing1ID,
		Name:        "name",
	}
	badEvent3 := &thing.ThingValue{
		PublisherID: publisherID,
		ThingID:     thing1ID,
		Name:        "baddate",
		Created:     "notadate",
	}
	badEvent4 := &thing.ThingValue{
		PublisherID: publisherID,
		ThingID:     "", // missing ID
		Name:        "temperature",
	}
	propsList := make(map[string][]byte)
	propsList[vocab.PropNameBattery] = []byte("50")
	propsList[vocab.PropNameCPULevel] = []byte("30")
	propsList[vocab.PropNameSwitch] = []byte("off")
	propsValue, _ := json.Marshal(propsList)
	props1 := &thing.ThingValue{
		PublisherID: publisherID,
		ThingID:     thing1ID,
		Name:        history.EventNameProperties,
		ValueJSON:   propsValue,
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
	err = addHist.AddAction(ctx, badEvent1)
	assert.Error(t, err)
	err = addHist.AddAction(ctx, nil)
	assert.Error(t, err)

	// verify named properties from different sources
	props := readHist.GetProperties(ctx, []string{vocab.PropNameTemperature, vocab.PropNameSwitch})
	assert.Equal(t, 2, len(props))
	assert.Equal(t, vocab.PropNameTemperature, props[0].Name)
	assert.Equal(t, []byte(temp1), props[0].ValueJSON)
	assert.Equal(t, vocab.PropNameSwitch, props[1].Name)

	// restart
	readHist.Release()
	addHist.Release()
	closeFn()

	backend := cmd.NewBucketStore(testFolder, testClientID, HistoryStoreBackend)
	cfg := config.NewHistoryConfig(testFolder)
	svc := service.NewHistoryService(&cfg, backend, nil)
	err = svc.Start()
	assert.NoError(t, err)

	// after closing and reopen the store the properties should still be there
	readHist, _ = svc.CapReadHistory(ctx, clientID, publisherID, thing1ID)
	props = readHist.GetProperties(ctx, []string{vocab.PropNameTemperature, vocab.PropNameSwitch})
	assert.Equal(t, 2, len(props))
	assert.Equal(t, props[0].Name, vocab.PropNameTemperature)
	assert.Equal(t, props[0].ValueJSON, []byte(temp1))
	assert.Equal(t, props[1].Name, vocab.PropNameSwitch)
	readHist.Release()

	err = svc.Stop()
	assert.NoError(t, err)
}

func TestGetLatest(t *testing.T) {
	logrus.Info("--- TestGetLatest ---")
	const count = 1000
	const clientID = "client1"
	const publisherID = "device1"
	const thing1ID = thingIDPrefix + "0" // matches a percentage of the random things
	store, closeFn := newHistoryService(useTestCapnp)
	defer closeFn()

	ctx := context.Background()

	// 10 sensors -> 1 sample per minute, 60 per hour -> 600
	// TODO: use different timezones
	highestFromAdded := addHistory(store, count, 1, 3600*24*30)

	readHistory, _ := store.CapReadHistory(ctx, clientID, publisherID, thing1ID)
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
	logrus.Info("--- TestPrevNext ---")
	const count = 1000
	const clientID = "client1"
	const thing0ID = thingIDPrefix + "0" // matches a percentage of the random things
	const publisherID = "device1"
	store, closeFn := newHistoryService(useTestCapnp)
	defer closeFn()

	ctx := context.Background()

	// 10 sensors -> 1 sample per minute, 60 per hour -> 600
	// TODO: use different timezones
	_ = addHistory(store, count, 1, 3600*24*30)

	readHistory, _ := store.CapReadHistory(ctx, clientID, publisherID, thing0ID)
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
	logrus.Info("--- TestPrevNextFiltered ---")
	const count = 1000
	const publisherID = "device1"
	const thing0ID = thingIDPrefix + "0" // matches a percentage of the random things
	const clientID = "client1"
	store, closeFn := newHistoryService(useTestCapnp)
	defer closeFn()

	ctx := context.Background()

	// 10 sensors -> 1 sample per minute, 60 per hour -> 600
	// TODO: use different timezones
	_ = addHistory(store, count, 1, 3600*24*30)
	propName := names[2] // names used to generate the history

	readHistory, _ := store.CapReadHistory(ctx, clientID, publisherID, thing0ID)
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
	logrus.Info("--- TestGetInfo ---")
	const publisherID = "device1"
	const thing0ID = thingIDPrefix + "0"

	store, stopFn := newHistoryService(useTestCapnp)
	defer stopFn()
	addHistory(store, 1000, 5, 1000)
	ctx := context.Background()

	//info := store.Info(ctx)
	//t.Logf("Store ID:%s, records:%d", info.Id, info.NrRecords)

	readHistory, _ := store.CapReadHistory(ctx, "test", publisherID, thing0ID)
	defer readHistory.Release()

	info := readHistory.Info(ctx)
	assert.NotEmpty(t, info.Engine)
	assert.NotEmpty(t, info.Id)
	t.Logf("ID:%s records:%d", info.Id, info.NrRecords)
}

func TestPubSub(t *testing.T) {
	const publisherID = "device1"
	const thing0ID = thingIDPrefix + "0"

	logrus.Info("--- TestPubSub ---")
	// start the pubsub server
	ctx := context.Background()
	svcConfig := config.NewHistoryConfig(testFolder)
	svcConfig.Retention = []history.EventRetention{
		{Name: vocab.PropNameTemperature},
		{Name: vocab.PropNameBattery},
	}
	pubSubSvc := service2.NewPubSubService()
	err := pubSubSvc.Start()
	require.NoError(t, err)
	// get the pubsub client for the history service
	psClient, err := pubSubSvc.CapServicePubSub(ctx, svcConfig.ServiceID)
	require.NoError(t, err)

	// create a new empty store to use
	_ = os.RemoveAll(svcConfig.Directory)
	store := cmd.NewBucketStore(testFolder, testClientID, HistoryStoreBackend)

	// start the service using pubsub
	svc := service.NewHistoryService(&svcConfig, store, psClient)
	err = svc.Start()
	if err != nil {
		logrus.Fatalf("Failed starting the state service: %s", err)
	}

	// publish events
	names := []string{
		vocab.PropNameTemperature, vocab.PropNameSwitch,
		vocab.PropNameSwitch, vocab.PropNameBattery,
		vocab.PropNameAlarm, "noname",
		"tttt", vocab.PropNameTemperature,
		vocab.PropNameSwitch, vocab.PropNameTemperature}
	_ = names

	// only valid names should be added
	for i := 0; i < 10; i++ {
		err = psClient.PubEvent(ctx, thing0ID, names[i], []byte("0.3"))
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 3)
	}

	time.Sleep(time.Millisecond * 100)
	// read back
	readHistory, _ := svc.CapReadHistory(ctx, "test", svcConfig.ServiceID, thing0ID)
	cursor := readHistory.GetEventHistory(ctx, "")
	ev, valid := cursor.First()
	assert.True(t, valid)
	assert.NotEmpty(t, ev)
	batched, _ := cursor.NextN(10)
	// expect 4 entries total from valid events
	assert.Equal(t, 3, len(batched))
	cursor.Release()

	// cleanup
	readHistory.Release()
	err = svc.Stop()
	assert.NoError(t, err)
	err = pubSubSvc.Stop()
	assert.NoError(t, err)
}

func TestManageRetention(t *testing.T) {
	logrus.Info("--- TestManageRetention ---")
	const publisherID = "device1"
	const thing0ID = thingIDPrefix + "0"

	svc, stopFn := newHistoryService(useTestCapnp)
	addHistory(svc, 1000, 5, 1000)
	ctx := context.Background()

	//info := store.Info(ctx)
	//t.Logf("Store ID:%s, records:%d", info.Id, info.NrRecords)
	mr, err := svc.CapManageRetention(ctx, "testclient")
	require.NoError(t, err)

	// verify the default retention list
	retList, err := mr.GetEvents(ctx)
	require.NoError(t, err)
	assert.Greater(t, 1, len(retList))

	// add a couple of retention
	newRet := history.EventRetention{Name: vocab.PropNameTemperature}
	err = mr.SetEventRetention(ctx, newRet)
	newRet = history.EventRetention{
		Name: "blob1", Publishers: []string{publisherID}}
	err = mr.SetEventRetention(ctx, newRet)
	require.NoError(t, err)

	// The new retention should now exist
	retList2, err := mr.GetEvents(ctx)
	require.NoError(t, err)
	assert.Equal(t, len(retList)+2, len(retList2))
	ret3, err := mr.GetEventRetention(ctx, "blob1")
	require.NoError(t, err)
	assert.Equal(t, "blob1", ret3.Name)
	valid, err := mr.TestEvent(ctx, &thing.ThingValue{
		PublisherID: publisherID,
		ThingID:     thing0ID,
		Name:        "blob1",
	})
	assert.NoError(t, err)
	assert.True(t, valid)

	// events of blob1 should be accepted now
	addHist, _ := svc.CapAddHistory(ctx, "testclient", false)
	err = addHist.AddEvent(ctx, thing.NewThingValue(
		publisherID, thing0ID, "blob1", []byte("hi)")))
	assert.NoError(t, err)
	//
	readHistory, _ := svc.CapReadHistory(ctx, "test", publisherID, thing0ID)
	cursor := readHistory.GetEventHistory(ctx, "blob1")
	histEv, valid := cursor.First()
	assert.True(t, valid)
	assert.Equal(t, "blob1", histEv.Name)

	//
	err = mr.RemoveEventRetention(ctx, "blob1")
	assert.NoError(t, err)

	// cleanup
	addHist.Release()
	readHistory.Release()
	mr.Release()
	stopFn()
}
