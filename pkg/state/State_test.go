package state_test

import (
	"context"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub/pkg/state"
	"github.com/hiveot/hub/pkg/state/capnpclient"
	"github.com/hiveot/hub/pkg/state/capnpserver"
	"github.com/hiveot/hub/pkg/state/service/statekvstore"
)

const stateStoreFile = "/tmp/statestore_test.json"
const testAddress = "/tmp/statestore_test.socket"
const testUseCapnp = true

// return an API to the state store, optionally using capnp RPC
func createStateStore(useCapnp bool) (store state.IState, stopFn func(), err error) {
	_ = os.Remove(stateStoreFile)
	stateStore, err := statekvstore.NewStateKVStore(stateStoreFile)

	// optionally test with capnp RPC
	if err == nil && useCapnp {
		ctx, cancelCtx := context.WithCancel(context.Background())
		_ = syscall.Unlink(testAddress)
		lis, _ := net.Listen("unix", testAddress)
		go func() {
			err = capnpserver.StartStateCapnpServer(ctx, lis, stateStore)
		}()

		capClient, err := capnpclient.NewStateCapnpClient(testAddress, true)
		// the stop function cancels the context, closes the listener and stops the store
		return capClient, func() {
			cancelCtx()
			_ = lis.Close()
			stateStore.Stop()
		}, err
	}
	return stateStore, stateStore.Stop, err
}

func TestStartStop(t *testing.T) {
	store, stopFn, err := createStateStore(testUseCapnp)
	require.NoError(t, err)
	assert.NotNil(t, store)

	stopFn()
}

func TestGetSet(t *testing.T) {
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	const val1 = "value 1"
	ctx := context.Background()
	store, stopFn, err := createStateStore(testUseCapnp)
	assert.NoError(t, err)
	clientStore := store.CapClientState(ctx, clientID1, appID)

	err = clientStore.Set(ctx, key1, val1)
	assert.NoError(t, err)

	val2, err := clientStore.Get(ctx, key1)
	assert.NoError(t, err)
	assert.Equal(t, val1, val2)

	// check if it persists
	//store, stopFn, err = createStateStore(testUseCapnp)
	clientStore2 := store.CapClientState(ctx, clientID1, appID)
	val3, err := clientStore2.Get(ctx, key1)
	assert.NoError(t, err)
	assert.Equal(t, val1, val3)

	stopFn()
}

func TestGetDifferentClient(t *testing.T) {
	const clientID1 = "test-client1"
	const clientID2 = "test-client2"
	const appID = "test-app"
	const key1 = "key1"
	const val1 = "value 1"
	ctx := context.Background()
	store, stopFn, err := createStateStore(testUseCapnp)
	assert.NoError(t, err)
	clientStore := store.CapClientState(ctx, clientID1, appID)

	err = clientStore.Set(ctx, key1, val1)
	assert.NoError(t, err)

	clientStore2 := store.CapClientState(ctx, clientID2, appID)
	val2, err := clientStore2.Get(ctx, key1)
	assert.Error(t, err)
	assert.NotEqual(t, val1, val2)
	stopFn()
}
