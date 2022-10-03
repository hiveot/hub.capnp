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
func createStateStore(useCapnp bool) (state.IState, error) {
	_ = os.Remove(stateStoreFile)
	store, _ := statekvstore.NewStateKVStoreServer(stateStoreFile)

	// optionally test with capnp RPC
	if useCapnp {
		ctx := context.Background()
		_ = syscall.Unlink(testAddress)
		lis, _ := net.Listen("unix", testAddress)
		go capnpserver.StartStateCapnpServer(ctx, lis, store)

		capClient, err := capnpclient.NewStateCapnpClient(testAddress, true)
		return capClient, err
	}
	return store, nil
}

func TestStartStop(t *testing.T) {
	store, err := createStateStore(testUseCapnp)
	require.NoError(t, err)
	assert.NotNil(t, store)

	//store.Stop()
}

func TestGetSet(t *testing.T) {
	const key1 = "key1"
	const val1 = "value 1"
	ctx := context.Background()
	store, err := createStateStore(testUseCapnp)
	assert.NoError(t, err)

	err = store.Set(ctx, key1, val1)
	assert.NoError(t, err)

	val2, err := store.Get(ctx, key1)
	assert.NoError(t, err)
	assert.Equal(t, val1, val2)

	//store.Stop()
}
