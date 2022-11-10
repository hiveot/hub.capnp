package state_test

import (
	"context"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/pkg/state"
	"github.com/hiveot/hub/pkg/state/capnpclient"
	"github.com/hiveot/hub/pkg/state/capnpserver"
	"github.com/hiveot/hub/pkg/state/config"
	"github.com/hiveot/hub/pkg/state/service"
)

const storeDir = "/tmp/test-state"
const stateStoreFile = "/tmp/statestore_test.json"
const testAddress = "/tmp/statestore_test.socket"
const testUseCapnp = true

//var backend = config.StateBackendKVStore

var backend = config.StateBackendBBolt

//var backend = config.StateBackendPebble

// return an API to the state service, optionally using capnp RPC
func createStateService(useCapnp bool) (store state.IStateService, stopFn func() error, err error) {
	_ = os.Remove(stateStoreFile)

	os.RemoveAll(storeDir)
	cfg := config.NewStateConfig(storeDir)
	cfg.Backend = backend
	stateSvc := service.NewStateStoreService(cfg)
	err = stateSvc.Start()

	// optionally test with capnp RPC
	if err == nil && useCapnp {

		ctx, cancelCtx := context.WithCancel(context.Background())

		_ = syscall.Unlink(testAddress)
		srvListener, _ := net.Listen("unix", testAddress)
		go func() {
			_ = capnpserver.StartStateCapnpServer(ctx, srvListener, stateSvc)
		}()
		// connect the client to the server above
		clConn, _ := net.Dial("unix", testAddress)
		capClient, err2 := capnpclient.NewStateCapnpClient(ctx, clConn)
		// the stop function cancels the context, closes the listener and stops the store
		return capClient, func() error {
			// don't kill the capnp messenger yet as capabilities are being released in the test cases
			time.Sleep(time.Millisecond)

			err = capClient.Stop()
			//_ = clConn.Close()
			cancelCtx()
			err = stateSvc.Stop()
			return err
		}, err2
	}
	return stateSvc, func() error { return stateSvc.Stop() }, err
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")
	store, stopFn, err := createStateService(testUseCapnp)
	require.NoError(t, err)
	assert.NotNil(t, store)

	err = stopFn()
	assert.NoError(t, err)
}

func TestStartStopBadLocation(t *testing.T) {
	logrus.Infof("--- TestStartStopBadLocation ---")
	// read-only folder
	cfg := config.NewStateConfig(storeDir)
	cfg.Backend = backend
	cfg.StoreDirectory = "/var"
	stateSvc := service.NewStateStoreService(cfg)
	err := stateSvc.Start()
	if !assert.Error(t, err) {
		stateSvc.Stop()
	}

	// stop twice should return an error
	err = stateSvc.Stop()
	assert.Error(t, err)

	// not a folder
	cfg.StoreDirectory = "/not/a/folder"
	stateSvc = service.NewStateStoreService(cfg)
	err = stateSvc.Start()
	if !assert.Error(t, err) {
		stateSvc.Stop()
	}
}

func TestSetGet(t *testing.T) {
	logrus.Infof("--- TestSetGet ---")
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	var val1 = []byte("value 1")

	ctx := context.Background()
	store, stopFn, err := createStateService(testUseCapnp)
	require.NoError(t, err)
	clientState, err := store.CapClientBucket(ctx, clientID1, appID)
	assert.NoError(t, err)

	err = clientState.Set(ctx, key1, val1)
	assert.NoError(t, err)

	val2, err := clientState.Get(ctx, key1)
	assert.NoError(t, err)
	assert.Equal(t, val1, val2)

	// check if it persists
	clientState2, err2 := store.CapClientBucket(ctx, clientID1, appID)
	assert.NoError(t, err2)
	val3, err := clientState2.Get(ctx, key1)
	assert.NoError(t, err)
	assert.Equal(t, val1, val3)

	clientState.Release()
	//clientState2.Release()
	stopFn()
}

func TestSetGetMultiple(t *testing.T) {
	logrus.Infof("--- TestSetGetMultiple ---")
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	const key2 = "key2"
	var val1 = []byte("value 1")
	var val2 = []byte("value 2")
	data := map[string][]byte{
		key1: val1,
		key2: val2,
	}

	ctx := context.Background()
	store, stopFn, err := createStateService(testUseCapnp)
	clientState, err := store.CapClientBucket(ctx, clientID1, appID)

	// write multiple
	err = clientState.SetMultiple(ctx, data)
	assert.NoError(t, err)

	// result must match
	data2, err := clientState.GetMultiple(ctx, []string{key1, key2})
	_ = data2
	assert.NoError(t, err)
	assert.Equal(t, data[key1], data2[key1])

	// cleanup
	clientState.Release()
	err = stopFn()
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	logrus.Infof("--- TestDelete ---")
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	var val1 = []byte("value 1")

	ctx := context.Background()
	store, stopFn, err := createStateService(testUseCapnp)
	require.NoError(t, err)
	clientState, err := store.CapClientBucket(ctx, clientID1, appID)
	if assert.NoError(t, err) {
		_ = clientState.Set(ctx, key1, val1)

		err = clientState.Delete(ctx, key1)
		assert.NoError(t, err)
		val2, _ := clientState.Get(ctx, key1)
		assert.Nil(t, val2)

		clientState.Release()
	}
	defer stopFn()
	//assert.NoError(t, err)
}

func TestGetDifferentClientBuckets(t *testing.T) {
	logrus.Infof("--- TestGetDifferentClientBuckets ---")
	const clientID1 = "test-client1"
	const clientID2 = "test-client2"
	const appID = "test-app"
	const key1 = "key1"
	var val1 = []byte("value 1")

	ctx := context.Background()
	store, stopFn, err := createStateService(testUseCapnp)
	assert.NoError(t, err)
	clientState, err := store.CapClientBucket(ctx, clientID1, appID)
	assert.NoError(t, err)

	err = clientState.Set(ctx, key1, val1)
	assert.NoError(t, err)
	clientState.Release()

	// second client
	clientStore2, err2 := store.CapClientBucket(ctx, clientID2, appID)
	assert.NoError(t, err2)
	val2, err := clientStore2.Get(ctx, key1)
	assert.NotEqual(t, val1, val2)
	clientStore2.Release()

	// we want to detect exceptions so don't use defer stopFn() which can hang on a lock
	stopFn()
}

func TestCursor(t *testing.T) {
	logrus.Infof("--- TestCursor ---")
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	const key2 = "key2"
	var val1 = []byte("value 1")
	var val2 = []byte("value 2")
	data := map[string][]byte{
		key1: val1,
		key2: val2,
	}

	ctx := context.Background()
	svc, stopFn, err := createStateService(testUseCapnp)
	clientState, err := svc.CapClientBucket(ctx, clientID1, appID)

	// write multiple
	err = clientState.SetMultiple(ctx, data)
	assert.NoError(t, err)

	// result must match
	cursor := clientState.Cursor(ctx)
	assert.NotNil(t, cursor)
	k1, _ := cursor.First()
	assert.NotNil(t, k1)
	k0, _ := cursor.Prev()
	assert.Empty(t, k0)
	k2, v := cursor.Seek(key1)
	assert.Equal(t, key1, k2)
	assert.Equal(t, val1, v)
	k3, _ := cursor.Next()
	assert.Equal(t, key2, k3)
	k4, _ := cursor.Last()
	assert.Equal(t, key2, k4)
	//
	cursor.Release()

	// cleanup
	clientState.Release()

	err = stopFn()
	assert.NoError(t, err)
}
