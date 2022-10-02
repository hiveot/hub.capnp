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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
func createNewStore(useCapnp bool) (directory.IDirectory, error) {
	_ = os.Remove(dirStoreFile)
	store, _ := directorykvstore.NewDirectoryKVStoreServer(dirStoreFile)

	// optionally test with capnp RPC
	if useCapnp {
		ctx := context.Background()
		_ = syscall.Unlink(testAddress)
		lis, _ := net.Listen("unix", testAddress)
		go capnpserver.StartDirectoryCapnpServer(ctx, lis, store)

		capClient, err := capnpclient.NewDirectoryStoreCapnpClient(testAddress, true)
		return capClient, err
	}
	return store, nil
}

func createTDDoc(thingID string, title string) string {
	td := &thing.ThingDescription{
		ID:    thingID,
		Title: title,
	}
	tdDoc, _ := json.Marshal(td)
	return string(tdDoc)
}

func TestStartStop(t *testing.T) {
	_ = os.Remove(dirStoreFile)
	store, err := createNewStore(testUseCapnp)
	require.NoError(t, err)
	assert.NotNil(t, store)
}

func TestAddRemoveTD(t *testing.T) {
	_ = os.Remove(dirStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	store, err := createNewStore(testUseCapnp)
	require.NoError(t, err)
	readCap := store.CapReadDirectory()
	updateCap := store.CapUpdateDirectory()

	ctx := context.Background()
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
}

func TestListTDs(t *testing.T) {
	_ = os.Remove(dirStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	store, err := createNewStore(testUseCapnp)
	require.NoError(t, err)
	readCap := store.CapReadDirectory()
	updateCap := store.CapUpdateDirectory()
	tdDoc1 := createTDDoc(thing1ID, title1)

	ctx := context.Background()
	err = updateCap.UpdateTD(ctx, thing1ID, tdDoc1)
	require.NoError(t, err)

	tdList, err := readCap.ListTDs(ctx, 0, 0)
	require.NoError(t, err)
	assert.NotNil(t, tdList)
	assert.True(t, len(tdList) > 0)
}

func TestQueryTDs(t *testing.T) {
	_ = os.Remove(dirStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	store, err := createNewStore(testUseCapnp)
	require.NoError(t, err)
	readCap := store.CapReadDirectory()
	updateCap := store.CapUpdateDirectory()

	tdDoc1 := createTDDoc(thing1ID, title1)
	ctx := context.Background()
	err = updateCap.UpdateTD(ctx, thing1ID, tdDoc1)
	require.NoError(t, err)

	jsonPathQuery := `$[?(@.id=="thing1")]`
	tdList, err := readCap.QueryTDs(ctx, jsonPathQuery, 0, 0)
	require.NoError(t, err)
	assert.NotNil(t, tdList)
	assert.True(t, len(tdList) > 0)
	el0 := thing.ThingDescription{}
	json.Unmarshal([]byte(tdList[0]), &el0)
	assert.Equal(t, thing1ID, el0.ID)
	assert.Equal(t, title1, el0.Title)
}

// simple performance test update/read, comparing direct vs capnp access
func TestPerf(t *testing.T) {
	_ = os.Remove(dirStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	const count = 1000

	store, err := createNewStore(true)
	require.NoError(t, err)
	readCap := store.CapReadDirectory()
	updateCap := store.CapUpdateDirectory()
	ctx := context.Background()

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

}
