package main_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/pkg/directorystore/thingkvstore"
)

const thingStoreFile = "/tmp/thingstore_test.json"

// create a store.
func createNewStore() *thingkvstore.ThingKVStoreServer {
	_ = os.Remove(thingStoreFile)
	store, _ := thingkvstore.NewThingKVStoreServer(thingStoreFile)
	return store
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
	tdDoc1 := createTDDoc(thing1ID, title1)
	err := store.UpdateTD(ctx, thing1ID, string(tdDoc1))
	require.NoError(t, err)
	assert.NotNil(t, store)

	td2, err := store.GetTD(ctx, thing1ID)
	require.NoError(t, err)
	assert.NotNil(t, td2)
	assert.Equal(t, tdDoc1, td2)

	err = store.RemoveTD(ctx, thing1ID)
	require.NoError(t, err)
	td3, err := store.GetTD(ctx, thing1ID)
	require.Error(t, err)
	assert.Equal(t, "", td3)
}

func TestListTDs(t *testing.T) {
	_ = os.Remove(thingStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	store := createNewStore()
	assert.NotNil(t, store)
	tdDoc1 := createTDDoc(thing1ID, title1)

	ctx := context.Background()
	err := store.UpdateTD(ctx, thing1ID, tdDoc1)
	require.NoError(t, err)
	assert.NotNil(t, store)

	tdList, err := store.ListTDs(ctx, 0, 0, nil)
	require.NoError(t, err)
	assert.NotNil(t, tdList)
	assert.True(t, len(tdList) > 0)
}

func TestQueryTDs(t *testing.T) {
	_ = os.Remove(thingStoreFile)
	const thing1ID = "thing1"
	const title1 = "title1"
	store := createNewStore()
	assert.NotNil(t, store)

	tdDoc1 := createTDDoc(thing1ID, title1)
	ctx := context.Background()
	err := store.UpdateTD(ctx, thing1ID, tdDoc1)
	require.NoError(t, err)
	assert.NotNil(t, store)

	jsonPathQuery := `$[?(@.id=="thing1")]`
	tdList, err := store.QueryTDs(ctx, jsonPathQuery, 0, 0, nil)
	require.NoError(t, err)
	assert.NotNil(t, tdList)
	assert.True(t, len(tdList) > 0)
	el0 := thing.ThingDescription{}
	json.Unmarshal([]byte(tdList[0]), &el0)
	assert.Equal(t, thing1ID, el0.ID)
	assert.Equal(t, title1, el0.Title)
}
