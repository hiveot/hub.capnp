package directory_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/resolver"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/bucketstore/kvbtree"
	"github.com/hiveot/hub/pkg/pubsub"
	service2 "github.com/hiveot/hub/pkg/pubsub/service"

	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpclient"
	"github.com/hiveot/hub/pkg/directory/capnpserver"
	"github.com/hiveot/hub/pkg/directory/service"
)

var testFolder = path.Join(os.TempDir(), "test-directory")
var testStoreFile = path.Join(testFolder, "directory.json")

const testUseCapnp = true

// startDirectory initializes a Directory service, optionally using capnp RPC
func startDirectory(useCapnp bool, svcPubSub pubsub.IServicePubSub) (dir directory.IDirectory, stopFn func()) {

	logrus.Infof("startDirectory start")
	defer logrus.Infof("startDirectory ended")
	_ = os.Remove(testStoreFile)
	store := kvbtree.NewKVStore(directory.ServiceName, testStoreFile)
	err := store.Open()
	if err != nil {
		panic("unable to open directory store")
	}
	svc := service.NewDirectoryService("", store, svcPubSub)
	err = svc.Start()
	if err != nil {
		panic("service fails to start")
	}

	// optionally test with capnp RPC
	if useCapnp {
		// start the server
		srvListener, err := net.Listen("tcp", ":0")
		if err != nil {
			logrus.Panic("Unable to create a listener, can't run test")
		}
		go capnpserver.StartDirectoryServiceCapnpServer(svc, srvListener)

		// connect the client to the server above
		capClient, _ := hubclient.ConnectWithCapnpTCP(srvListener.Addr().String(), nil, nil)
		dirClient := capnpclient.NewDirectoryCapnpClient(capClient)
		return dirClient, func() {
			dirClient.Release()
			_ = srvListener.Close()
			_ = svc.Stop()
			_ = store.Close()
		}
	}
	return svc, func() {
		_ = svc.Stop()
		_ = store.Close()
	}
}

// generate a JSON serialized TD document
func createTDDoc(thingID string, title string) []byte {
	td := &thing.TD{
		ID:    thingID,
		Title: title,
	}
	tdDoc, _ := json.Marshal(td)
	return tdDoc
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	// clean start
	_ = os.RemoveAll(testFolder)
	_ = os.MkdirAll(testFolder, 0700)

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop start ---")

	_ = os.Remove(testStoreFile)
	store, stopFunc := startDirectory(testUseCapnp, nil)
	defer stopFunc()
	assert.NotNil(t, store)
	logrus.Infof("--- TestStartStop end ---")
}

func TestAddRemoveTD(t *testing.T) {
	logrus.Infof("--- TestAddRemoveTD start ---")
	_ = os.Remove(testStoreFile)
	const publisherID = "urn:test"
	const thing1ID = "urn:thing1"
	const title1 = "title1"
	ctx := context.Background()
	store, stopFunc := startDirectory(testUseCapnp, nil)
	defer stopFunc()

	readCap, err := store.CapReadDirectory(ctx, thing1ID)
	assert.NoError(t, err)
	updateCap, err := store.CapUpdateDirectory(ctx, thing1ID)
	assert.NoError(t, err)
	require.NotNil(t, readCap)
	require.NotNil(t, updateCap)

	tdDoc1 := createTDDoc(thing1ID, title1)
	err = updateCap.UpdateTD(ctx, publisherID, thing1ID, tdDoc1)
	assert.NoError(t, err)

	tv2, err := readCap.GetTD(ctx, publisherID, thing1ID)
	if assert.NoError(t, err) {
		assert.NotNil(t, tv2)
		assert.Equal(t, thing1ID, tv2.ThingID)
		assert.Equal(t, tdDoc1, tv2.Data)
	}
	err = updateCap.RemoveTD(ctx, publisherID, thing1ID)
	assert.NoError(t, err)
	td3, err := readCap.GetTD(ctx, publisherID, thing1ID)
	assert.Empty(t, td3)
	assert.Error(t, err)

	readCap.Release()
	updateCap.Release()

	logrus.Infof("--- TestRemoveTD end ---")
}

//func TestListTDs(t *testing.T) {
//	logrus.Infof("--- TestListTDs start ---")
//	_ = os.Remove(dirStoreFile)
//	const thing1ID = "thing1"
//	const title1 = "title1"
//
//	ctx := context.Background()
//	store, cancelFunc, err := startDirectory(testUseCapnp)
//	defer cancelFunc()
//	require.NoError(t, err)
//
//	readCap := store.CapReadDirectory(ctx)
//	defer readCap.Release()
//	updateCap := store.CapUpdateDirectory(ctx)
//	defer updateCap.Release()
//	tdDoc1 := createTDDoc(thing1ID, title1)
//
//	err = updateCap.UpdateTD(ctx, thing1ID, tdDoc1)
//	require.NoError(t, err)
//
//	tdList, err := readCap.ListTDs(ctx, 0, 0)
//	require.NoError(t, err)
//	assert.NotNil(t, tdList)
//	assert.True(t, len(tdList) > 0)
//	logrus.Infof("--- TestListTDs end ---")
//}

func TestCursor(t *testing.T) {
	logrus.Infof("--- TestCursor start ---")
	_ = os.Remove(testStoreFile)
	const publisherID = "urn:test"
	const thing1ID = "urn:thing1"
	const title1 = "title1"

	ctx := context.Background()
	store, stopFunc := startDirectory(testUseCapnp, nil)
	defer stopFunc()

	readCap, err := store.CapReadDirectory(ctx, thing1ID)
	assert.NoError(t, err)
	defer readCap.Release()
	updateCap, err := store.CapUpdateDirectory(ctx, thing1ID)
	assert.NoError(t, err)
	defer updateCap.Release()

	// add 1 doc. the service itself also has a doc
	tdDoc1 := createTDDoc(thing1ID, title1)
	err = updateCap.UpdateTD(ctx, publisherID, thing1ID, tdDoc1)
	require.NoError(t, err)

	// expect 2 docs, the service itself and the one just added
	cursor := readCap.Cursor(ctx)
	assert.NoError(t, err)
	defer cursor.Release()

	tdValue, valid := cursor.First()
	assert.True(t, valid)
	assert.NotEmpty(t, tdValue)
	assert.NotEmpty(t, tdValue.Data)

	tdValue, valid = cursor.Next() // second
	assert.True(t, valid)
	assert.NotEmpty(t, tdValue)
	assert.NotEmpty(t, tdValue.Data)

	tdValue, valid = cursor.Next() // there is no third
	assert.False(t, valid)
	assert.Empty(t, tdValue)

	tdValues, valid := cursor.NextN(10) // still no third
	assert.False(t, valid)
	assert.Empty(t, tdValues)

	assert.NoError(t, err)
	logrus.Infof("--- TestCursor end ---")
}

func TestPubSub(t *testing.T) {
	logrus.Infof("--- TestPubSub start ---")
	_ = os.Remove(testStoreFile)
	const publisherID = "test"
	const thing1ID = "thing3"
	const title1 = "title1"
	ctx := context.Background()

	// use an in-memory version of the pubsub service
	pubSubSvc := service2.NewPubSubService()
	resolver.RegisterService[pubsub.IPubSubService](pubSubSvc)
	//cap := resolver.GetCapability[directory.IDirectory]()
	//cap.Release()

	err := pubSubSvc.Start()
	require.NoError(t, err)

	// get the pubsub capability for services
	svcPubSub, err := pubSubSvc.CapServicePubSub(ctx, "test")
	require.NoError(t, err)

	store, stopFunc := startDirectory(testUseCapnp, svcPubSub)
	defer stopFunc()

	// publish the TD
	tdDoc := createTDDoc(thing1ID, title1)
	err = svcPubSub.PubEvent(ctx, thing1ID, hubapi.EventNameTD, tdDoc)
	assert.NoError(t, err)

	// expect it to be added to the directory
	readCap, err := store.CapReadDirectory(ctx, thing1ID)
	assert.NoError(t, err)

	tv2, err := readCap.GetTD(ctx, publisherID, thing1ID)
	if assert.NoError(t, err) {
		assert.NotNil(t, tv2)
		assert.Equal(t, thing1ID, tv2.ThingID)
		assert.Equal(t, tdDoc, tv2.Data)
	}

	// cleanup
	svcPubSub.Release()
}

//func TestQueryTDs(t *testing.T) {
//	logrus.Infof("--- TestQueryTDs start ---")
//	_ = os.Remove(dirStoreFile)
//	const thing1ID = "thing1"
//	const title1 = "title1"
//
//	ctx := context.Background()
//	store, cancelFunc, err := startDirectory(testUseCapnp)
//	defer cancelFunc()
//	require.NoError(t, err)
//	readCap := store.CapReadDirectory(ctx)
//	defer readCap.Release()
//	updateCap := store.CapUpdateDirectory(ctx)
//	defer updateCap.Release()
//
//	tdDoc1 := createTDDoc(thing1ID, title1)
//	err = updateCap.UpdateTD(ctx, thing1ID, tdDoc1)
//	require.NoError(t, err)
//
//	jsonPathQuery := `$[?(@.id=="thing1")]`
//	tdList, err := readCap.QueryTDs(ctx, jsonPathQuery, 0, 0)
//	require.NoError(t, err)
//	assert.NotNil(t, tdList)
//	assert.True(t, len(tdList) > 0)
//	el0 := thing.ThingDescription{}
//	json.Unmarshal([]byte(tdList[0]), &el0)
//	assert.Equal(t, thing1ID, el0.ID)
//	assert.Equal(t, title1, el0.Title)
//	logrus.Infof("--- TestQueryTDs end ---")
//}

// simple performance test update/read, comparing direct vs capnp access
// TODO: turn into bench test
func TestPerf(t *testing.T) {
	logrus.Infof("--- start TestPerf ---")
	_ = os.Remove(testStoreFile)
	const publisherID = "urn:test"
	const thing1ID = "urn:thing1"
	const title1 = "title1"
	const count = 1000

	ctx := context.Background()
	store, stopFunc := startDirectory(true, nil)
	defer stopFunc()
	readCap, err := store.CapReadDirectory(ctx, thing1ID)
	assert.NoError(t, err)
	updateCap, err := store.CapUpdateDirectory(ctx, thing1ID)
	assert.NoError(t, err)

	// test update
	t1 := time.Now()
	for i := 0; i < count; i++ {
		tdDoc1 := createTDDoc(thing1ID, title1)
		err := updateCap.UpdateTD(ctx, publisherID, thing1ID, tdDoc1)
		require.NoError(t, err)
	}
	d1 := time.Now().Sub(t1)
	fmt.Printf("Duration for update %d iterations: %d msec\n", count, int(d1.Milliseconds()))

	// test read
	t2 := time.Now()
	for i := 0; i < count; i++ {
		td, err := readCap.GetTD(ctx, publisherID, thing1ID)
		require.NoError(t, err)
		assert.NotNil(t, td)
	}
	d2 := time.Now().Sub(t2)
	fmt.Printf("Duration for read %d iterations: %d msec\n", count, int(d2.Milliseconds()))

	readCap.Release()
	updateCap.Release()
	logrus.Infof("--- end TestPerf ---")
}
