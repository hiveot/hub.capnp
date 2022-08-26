package main_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/hub/pkg/svc/thingstore/thingkvstore"
	"github.com/wostzone/wost.grpc/go/svc"
	"github.com/wostzone/wost.grpc/go/thing"
)

const thingStoreFile = "/tmp/thingstore_test.json"

// Pre-condition: dapr must be initialized

func createNewStore() svc.ThingStoreServer {
	_ = os.Remove(thingStoreFile)
	store, _ := thingkvstore.NewThingKVStoreServer(thingStoreFile)
	return store
}

func TestStartStop(t *testing.T) {
	_ = os.Remove(thingStoreFile)
	store := createNewStore()
	assert.NotNil(t, store)
}

func TestAddRemoveTD(t *testing.T) {
	_ = os.Remove(thingStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	store := createNewStore()
	assert.NotNil(t, store)

	ctx := context.Background()
	td := &thing.ThingDescription{
		Id:    thing1ID,
		Title: title1,
	}
	_, err := store.Write(ctx, td)
	require.NoError(t, err)
	assert.NotNil(t, store)

	rargs := &svc.ReadTD_Args{ThingID: thing1ID}
	td2, err := store.Read(ctx, rargs)
	require.NoError(t, err)
	assert.NotNil(t, td2)
	assert.Equal(t, title1, td2.Title)

	_, err = store.Remove(ctx, &svc.RemoveTD_Args{ThingID: thing1ID})
	require.NoError(t, err)
	td3, err := store.Read(ctx, rargs)
	require.Error(t, err)
	assert.Nil(t, td3)
}

func TestListTDs(t *testing.T) {
	_ = os.Remove(thingStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	store := createNewStore()
	assert.NotNil(t, store)

	ctx := context.Background()
	td := &thing.ThingDescription{
		Id:    thing1ID,
		Title: title1,
	}
	_, err := store.Write(ctx, td)
	require.NoError(t, err)
	assert.NotNil(t, store)

	rargs := &svc.ListTD_Args{}
	tdList, err := store.List(ctx, rargs)
	require.NoError(t, err)
	assert.NotNil(t, tdList)
	assert.True(t, len(tdList.Things) > 0)
}

func TestQueryTDs(t *testing.T) {
	_ = os.Remove(thingStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	store := createNewStore()
	assert.NotNil(t, store)

	ctx := context.Background()
	td := &thing.ThingDescription{
		Id:    thing1ID,
		Title: title1,
	}
	_, err := store.Write(ctx, td)
	require.NoError(t, err)
	assert.NotNil(t, store)

	rargs := &svc.QueryTD_Args{
		JsonPathQuery: `$[?(@.id=="thing1")]`,
	}
	tdList, err := store.Query(ctx, rargs)
	require.NoError(t, err)
	assert.NotNil(t, tdList)
	assert.True(t, len(tdList.Things) > 0)
	assert.Equal(t, title1, tdList.Things[0].Title)
}
