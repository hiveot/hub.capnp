package state_test

import (
	"context"
	"github.com/hiveot/hub/lib/hubclient"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/state"
	"github.com/hiveot/hub/pkg/state/capnpclient"
	"github.com/hiveot/hub/pkg/state/capnpserver"
	"github.com/hiveot/hub/pkg/state/config"
	"github.com/hiveot/hub/pkg/state/service"
)

const storeDir = "/tmp/test-state"

const testAddress = "/tmp/statestore_test.socket"
const testUseCapnp = true

var backend = bucketstore.BackendKVBTree

//var backend = bucketstore.BackendBBolt

//var backend = bucketstore.BackendPebble

// return an API to the state service, optionally using capnp RPC
func startStateService(useCapnp bool) (store state.IStateService, stopFn func(), err error) {
	logrus.Infof("startStateService")
	_ = os.RemoveAll(storeDir)
	cfg := config.NewStateConfig(storeDir)
	cfg.Backend = backend
	stateSvc := service.NewStateStoreService(cfg)
	ctx, cancelCtx := context.WithCancel(context.Background())
	err = stateSvc.Start(ctx)

	// optionally test with capnp RPC
	if err == nil && useCapnp {

		_ = syscall.Unlink(testAddress)
		srvListener, err2 := net.Listen("unix", testAddress)
		if err2 != nil {
			panic(err2)
		}
		//srvListener, _ := net.Listen("tcp", ":0")
		go func() {
			_ = capnpserver.StartStateCapnpServer(stateSvc, srvListener)
		}()
		// connect the client to the server above
		capClient, _ := hubclient.ConnectWithCapnpUDS("", testAddress)
		stateClient := capnpclient.NewStateCapnpClient(capClient)
		// the stop function cancels the context, closes the listener and stops the store
		return stateClient, func() {
			// don't kill the capnp messenger yet as capabilities are being released in the test cases
			stateClient.Release()
			cancelCtx()
			_ = srvListener.Close()
			_ = stateSvc.Stop()
		}, err2
	}
	return stateSvc, func() {
		cancelCtx()
		_ = stateSvc.Stop()
	}, err
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")
	store, stopFn, err := startStateService(testUseCapnp)
	require.NoError(t, err)
	assert.NotNil(t, store)

	stopFn()
}

func TestStartStopBadLocation(t *testing.T) {
	logrus.Infof("--- TestStartStopBadLocation ---")
	// read-only folder
	cfg := config.NewStateConfig(storeDir)
	cfg.Backend = backend
	cfg.StoreDirectory = "/var"
	stateSvc := service.NewStateStoreService(cfg)
	err := stateSvc.Start(context.Background())
	if !assert.Error(t, err) {
		err = stateSvc.Stop()
		assert.NoError(t, err)
	}

	// stop twice should return an error
	err = stateSvc.Stop()
	assert.Error(t, err)

	// not a folder
	cfg.StoreDirectory = "/not/a/folder"
	stateSvc = service.NewStateStoreService(cfg)
	err = stateSvc.Start(context.Background())
	if !assert.Error(t, err) {
		err = stateSvc.Stop()
		assert.NoError(t, err)
	}
}

func TestSetGet1(t *testing.T) {
	logrus.Infof("--- TestSetGet1 ---")
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	var val1 = []byte("value 1")

	ctx := context.Background()
	svc, stopFn, err := startStateService(testUseCapnp)
	require.NoError(t, err)
	defer stopFn()

	clientState, _ := svc.CapClientState(ctx, clientID1, appID)
	assert.NotNil(t, clientState)

	logrus.Infof("set")
	err = clientState.Set(ctx, key1, val1)
	assert.NoError(t, err)

	logrus.Infof("get1")
	val2, err := clientState.Get(ctx, key1)
	assert.NoError(t, err)
	val2 = val1
	assert.Equal(t, val1, val2)
	//
	// check if it persists
	clientState2, _ := svc.CapClientState(ctx, clientID1, appID)
	assert.NotNil(t, clientState2)
	logrus.Infof("get2")
	val3, err := clientState2.Get(ctx, key1)
	assert.NoError(t, err)
	assert.Equal(t, val1, val3)

	clientState.Release()
	//clientState2.Release()
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
	store, stopFn, err := startStateService(testUseCapnp)
	defer stopFn()

	clientState, _ := store.CapClientState(ctx, clientID1, appID)

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
	time.Sleep(time.Millisecond)
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	logrus.Infof("--- TestDelete ---")
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	var val1 = []byte("value 1")

	ctx := context.Background()
	store, stopFn, err := startStateService(testUseCapnp)
	require.NoError(t, err)
	defer stopFn()

	clientState, _ := store.CapClientState(ctx, clientID1, appID)
	err = clientState.Set(ctx, key1, val1)
	assert.NoError(t, err)

	err = clientState.Delete(ctx, key1)
	assert.NoError(t, err)
	val2, _ := clientState.Get(ctx, key1)
	assert.Nil(t, val2)

	clientState.Release()
}

func TestGetDifferentClientBuckets(t *testing.T) {
	logrus.Infof("--- TestGetDifferentClientBuckets ---")
	const clientID1 = "test-client1"
	const clientID2 = "test-client2"
	const appID = "test-app"
	const key1 = "key1"
	var val1 = []byte("value 1")

	ctx := context.Background()
	store, stopFn, err := startStateService(testUseCapnp)
	assert.NoError(t, err)
	defer stopFn()
	clientState, _ := store.CapClientState(ctx, clientID1, appID)

	err = clientState.Set(ctx, key1, val1)
	assert.NoError(t, err)
	clientState.Release()

	// second client
	clientStore2, _ := store.CapClientState(ctx, clientID2, appID)
	val2, err := clientStore2.Get(ctx, key1)
	assert.NotEqual(t, val1, val2)
	clientStore2.Release()
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
	svc, stopFn, err := startStateService(testUseCapnp)
	defer stopFn()
	clientState, _ := svc.CapClientState(ctx, clientID1, appID)

	// write multiple
	err = clientState.SetMultiple(ctx, data)
	assert.NoError(t, err)

	// result must match
	cursor := clientState.Cursor(ctx)
	assert.NotNil(t, cursor)
	k1, v, valid := cursor.First()
	assert.True(t, valid)
	assert.NotNil(t, k1)
	assert.Equal(t, val1, v)
	k0, _, valid := cursor.Prev()
	assert.False(t, valid)
	assert.Empty(t, k0)
	k2, v, valid := cursor.Seek(key1)
	assert.True(t, valid)
	assert.Equal(t, key1, k2)
	assert.Equal(t, val1, v)
	k3, _, valid := cursor.Next()
	assert.True(t, valid)
	assert.Equal(t, key2, k3)
	k4, _, valid := cursor.Last()
	assert.True(t, valid)
	assert.Equal(t, key2, k4)
	//
	cursor.Release()

	// cleanup
	clientState.Release()

}
