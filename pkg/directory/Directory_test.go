package directory_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/directory/capnpclient"
	"github.com/hiveot/hub/pkg/directory/capnpserver"
	"github.com/hiveot/hub/pkg/directory/service/directorykvstore"
)

const dirStoreFile = "/tmp/directorystore_test.json"
const testAddress = "/tmp/dirstore_test.socket"
const testUseCapnp = true

// createNewStore returns an API to the directory store, optionally using capnp RPC
func createNewStore(useCapnp bool) (directory.IDirectory, func(), error) {

	ctx, cancelFunc := context.WithCancel(context.Background())
	logrus.Infof("createNewStore start")
	defer logrus.Infof("createNewStore ended")
	_ = os.Remove(dirStoreFile)
	store, _ := directorykvstore.NewDirectoryKVStoreServer(ctx, dirStoreFile)

	// optionally test with capnp RPC
	if useCapnp {
		// start the server
		_ = syscall.Unlink(testAddress)
		srvListener, err := net.Listen("unix", testAddress)
		if err != nil {
			logrus.Panic("Unable to create a listener, can't run test")
		}
		go capnpserver.StartDirectoryCapnpServer(ctx, srvListener, store)

		// connect the client to the server above
		clConn, _ := net.Dial("unix", testAddress)
		capClient, err := capnpclient.NewDirectoryCapnpClient(ctx, clConn)
		return capClient, func() { cancelFunc(); store.Stop(ctx) }, err
	}
	return store, func() { cancelFunc(); store.Stop(ctx) }, nil
}

func createTDDoc(thingID string, title string) string {
	td := &thing.ThingDescription{
		ID:    thingID,
		Title: title,
	}
	tdDoc, _ := json.Marshal(td)
	return string(tdDoc)
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop start ---")
	_ = os.Remove(dirStoreFile)
	store, cancelFunc, err := createNewStore(testUseCapnp)
	defer cancelFunc()
	require.NoError(t, err)
	assert.NotNil(t, store)
	logrus.Infof("--- TestStartStop end ---")
}

func TestAddRemoveTD(t *testing.T) {
	logrus.Infof("--- TestRemoveTD start ---")
	_ = os.Remove(dirStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	ctx := context.Background()
	store, cancelFunc, err := createNewStore(testUseCapnp)
	defer cancelFunc()
	require.NoError(t, err)
	readCap := store.CapReadDirectory(ctx)
	defer readCap.Release()
	updateCap := store.CapUpdateDirectory(ctx)
	defer updateCap.Release()

	tdDoc1 := createTDDoc(thing1ID, title1)
	err = updateCap.UpdateTD(ctx, thing1ID, string(tdDoc1))
	require.NoError(t, err)
	assert.NotNil(t, updateCap)

	td2, err := readCap.GetTD(ctx, thing1ID)
	require.NoError(t, err)
	assert.NotNil(t, td2)
	assert.Equal(t, tdDoc1, td2)

	err = updateCap.RemoveTD(ctx, thing1ID)
	require.NoError(t, err)
	td3, err := readCap.GetTD(ctx, thing1ID)
	require.Error(t, err)
	assert.Equal(t, "", td3)
	logrus.Infof("--- TestRemoveTD end ---")
}

func TestListTDs(t *testing.T) {
	logrus.Infof("--- TestListTDs start ---")
	_ = os.Remove(dirStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"

	ctx := context.Background()
	store, cancelFunc, err := createNewStore(testUseCapnp)
	defer cancelFunc()
	require.NoError(t, err)

	readCap := store.CapReadDirectory(ctx)
	defer readCap.Release()
	updateCap := store.CapUpdateDirectory(ctx)
	defer updateCap.Release()
	tdDoc1 := createTDDoc(thing1ID, title1)

	err = updateCap.UpdateTD(ctx, thing1ID, tdDoc1)
	require.NoError(t, err)

	tdList, err := readCap.ListTDs(ctx, 0, 0)
	require.NoError(t, err)
	assert.NotNil(t, tdList)
	assert.True(t, len(tdList) > 0)
	logrus.Infof("--- TestListTDs end ---")
}

func TestListTDcb(t *testing.T) {
	logrus.Infof("--- TestListTDcb start ---")
	_ = os.Remove(dirStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	const count = 1000
	tdList := make([]string, 0)

	ctx := context.Background()
	store, cancelFunc, err := createNewStore(testUseCapnp)
	defer cancelFunc()
	require.NoError(t, err)

	readCap := store.CapReadDirectory(ctx)
	defer readCap.Release()
	updateCap := store.CapUpdateDirectory(ctx)
	defer updateCap.Release()
	tdDoc1 := createTDDoc(thing1ID, title1)

	err = updateCap.UpdateTD(ctx, thing1ID, tdDoc1)
	t1 := time.Now()
	require.NoError(t, err)
	for i := 0; i < count && err == nil; i++ {
		err = readCap.ListTDcb(ctx, func(batch []string, isLast bool) error {
			tdList = append(tdList, batch...)
			return nil
		})
	}
	d1 := time.Now().Sub(t1)
	logrus.Infof("%d calls to ListTDcb: %d msec", count, d1.Milliseconds())
	require.NoError(t, err)
	assert.NotNil(t, tdList)
	assert.True(t, len(tdList) > 0)
	logrus.Infof("--- TestListTDcb end ---")
}

//func TestQueryTDs(t *testing.T) {
//	logrus.Infof("--- TestQueryTDs start ---")
//	_ = os.Remove(dirStoreFile)
//	const thing1ID = "thing1"
//	const title1 = "title1"
//
//	ctx := context.Background()
//	store, cancelFunc, err := createNewStore(testUseCapnp)
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
func TestPerf(t *testing.T) {
	logrus.Infof("--- start TestPerf ---")
	_ = os.Remove(dirStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	const count = 1000

	ctx := context.Background()
	store, cancelFunc, err := createNewStore(true)
	defer cancelFunc()
	require.NoError(t, err)
	readCap := store.CapReadDirectory(ctx)
	defer readCap.Release()
	updateCap := store.CapUpdateDirectory(ctx)
	defer updateCap.Release()

	// test update
	t1 := time.Now()
	for i := 0; i < count; i++ {
		tdDoc1 := createTDDoc(thing1ID, title1)
		//updateCap := store.CapUpdateDirectory()
		err = updateCap.UpdateTD(ctx, thing1ID, string(tdDoc1))
		require.NoError(t, err)
	}
	d1 := time.Now().Sub(t1)
	fmt.Printf("Duration for update %d iterations: %d msec\n", count, int(d1.Milliseconds()))

	// test read
	t2 := time.Now()
	for i := 0; i < count; i++ {
		td, err := readCap.GetTD(ctx, thing1ID)
		require.NoError(t, err)
		assert.NotNil(t, td)
	}
	d2 := time.Now().Sub(t2)
	fmt.Printf("Duration for read %d iterations: %d msec\n", count, int(d2.Milliseconds()))
	logrus.Infof("--- end TestPerf ---")
}
