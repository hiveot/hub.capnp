package thingstore_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/wost.grpc/go/svc"
	"github.com/wostzone/wost.grpc/go/thing"
	"svc/thingstore"
)

const testStoreName = "testStore"

// Pre-condition: dapr must be initialized

func TestStartStop(t *testing.T) {
	store, err := thingstore.NewThingStoreServer(testStoreName)
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestAddRemoveTD(t *testing.T) {
	const thing1ID = "thing1"
	const title1 = "title1"
	store, err := thingstore.NewThingStoreServer(testStoreName)
	require.NoError(t, err)

	ctx := context.Background()
	td := &thing.ThingDescription{
		Id:    thing1ID,
		Title: title1,
	}
	_, err = store.WriteTD(ctx, td)
	require.NoError(t, err)
	assert.NotNil(t, store)

	rargs := &svc.ReadTD_Args{ThingID: thing1ID}
	td2, err := store.ReadTD(ctx, rargs)
	require.NoError(t, err)
	assert.NotNil(t, td2)
	assert.Equal(t, title1, td2.Title)
}

func TestListTDs(t *testing.T) {
	const thing1ID = "thing1"
	const title1 = "title1"
	store, err := thingstore.NewThingStoreServer(testStoreName)
	require.NoError(t, err)

	ctx := context.Background()
	td := &thing.ThingDescription{
		Id:    thing1ID,
		Title: title1,
	}
	_, err = store.WriteTD(ctx, td)
	require.NoError(t, err)
	assert.NotNil(t, store)

	rargs := &svc.ListTD_Args{}
	tdList, err := store.ListTD(ctx, rargs)
	require.NoError(t, err)
	assert.NotNil(t, tdList)
	assert.True(t, len(tdList.Things) > 0)
}
