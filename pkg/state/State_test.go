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
	statekvstore "github.com/hiveot/hub/pkg/state/service"
)

const stateStoreFile = "/tmp/statestore_test.json"
const testAddress = "/tmp/statestore_test.socket"
const testUseCapnp = false

//
const testUseBBolt = false

// return an API to the state store, optionally using capnp RPC
func createStateStore(useCapnp bool) (store state.IStateStore, stopFn func() error, err error) {
	_ = os.Remove(stateStoreFile)

	ctx, cancelCtx := context.WithCancel(context.Background())

	cfg := config.NewStateConfig("/tmp/test-state")
	cfg.Backend = config.StateBackendKVStore
	if testUseBBolt {
		cfg.Backend = config.StateBackendBBolt
	}
	stateStore := statekvstore.NewStateStoreService(cfg)
	err = stateStore.Start()

	// optionally test with capnp RPC
	if err == nil && useCapnp {

		_ = syscall.Unlink(testAddress)
		srvListener, _ := net.Listen("unix", testAddress)
		go func() {
			_ = capnpserver.StartStateCapnpServer(ctx, srvListener, stateStore)
		}()
		// connect the client to the server above
		clConn, _ := net.Dial("unix", testAddress)
		capClient, err2 := capnpclient.NewStateCapnpClient(ctx, clConn)
		// the stop function cancels the context, closes the listener and stops the store
		return capClient, func() error {
			cancelCtx()
			_ = clConn.Close()
			err3 := stateStore.Stop()
			return err3
		}, err2
	}
	return stateStore, func() error { cancelCtx(); return stateStore.Stop() }, err
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	store, stopFn, err := createStateStore(testUseCapnp)
	require.NoError(t, err)
	assert.NotNil(t, store)

	err = stopFn()
	assert.NoError(t, err)
}

func TestGetSet(t *testing.T) {
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	const val1 = "value 1"

	ctx := context.Background()
	store, stopFn, err := createStateStore(testUseCapnp)
	require.NoError(t, err)
	defer stopFn()
	clientStore, err := store.CapClientState(ctx, clientID1, appID)
	assert.NoError(t, err)
	defer clientStore.Release()

	err = clientStore.Set(ctx, key1, val1)
	assert.NoError(t, err)

	val2, err := clientStore.Get(ctx, key1)
	assert.NoError(t, err)
	assert.Equal(t, val1, val2)

	// check if it persists
	//store, stopFn, err = createStateStore(testUseCapnp)
	clientStore2, err2 := store.CapClientState(ctx, clientID1, appID)
	assert.NoError(t, err2)
	val3, err := clientStore2.Get(ctx, key1)
	assert.NoError(t, err)
	assert.Equal(t, val1, val3)
}

func TestGetDifferentClient(t *testing.T) {
	const clientID1 = "test-client1"
	const clientID2 = "test-client2"
	const appID = "test-app"
	const key1 = "key1"
	const val1 = "value 1"
	ctx := context.Background()
	store, stopFn, err := createStateStore(testUseCapnp)
	defer stopFn()
	assert.NoError(t, err)
	clientStore, err := store.CapClientState(ctx, clientID1, appID)
	assert.NoError(t, err)

	err = clientStore.Set(ctx, key1, val1)
	assert.NoError(t, err)
	clientStore.Release()

	clientStore2, err2 := store.CapClientState(ctx, clientID2, appID)
	assert.NoError(t, err2)
	val2, err := clientStore2.Get(ctx, key1)
	clientStore2.Release()
	assert.Error(t, err)
	assert.NotEqual(t, val1, val2)
}

// test performance of N get/sets
func TestPerf(t *testing.T) {
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	const val1 = "value 1"
	const count = 1000

	ctx := context.Background()
	store, stopFn, err := createStateStore(testUseCapnp)
	require.NoError(t, err)
	defer stopFn()
	clientStore, err := store.CapClientState(ctx, clientID1, appID)
	assert.NoError(t, err)
	defer clientStore.Release()

	// update
	t1 := time.Now()
	for i := 0; i < count; i++ {
		err = clientStore.Set(ctx, key1, val1)
	}
	d1 := time.Now().Sub(t1)
	// read
	t2 := time.Now()
	for i := 0; i < count; i++ {
		_, _ = clientStore.Get(ctx, key1)
	}
	d2 := time.Now().Sub(t2)
	logrus.Infof("set '%d' times: %d msec", count, d1.Milliseconds())
	logrus.Infof("get '%d' times: %d msec", count, d2.Milliseconds())
}
