// Package dirstore with test cases for store implementations
package dirstore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Generic directory store testcases, invoked by specific implementation (eg dirfilestore)

func DirStoreStartStop(t *testing.T, store IDirStore) {
	err := store.Open()
	assert.NoError(t, err)
	store.Close()
}

func DirStoreCrud(t *testing.T, store IDirStore) {
	thingID := "thing1"
	thingTD1 := make(map[string]interface{})
	thingTD1["a"] = "this is a td"
	thingTD2 := make(map[string]interface{})
	thingTD2["b"] = "this is a td updated"
	err := store.Open()
	assert.NoError(t, err)
	// Create
	err = store.Replace(thingID, thingTD1)
	assert.NoError(t, err)
	// Read
	td2, err := store.Get(thingID)
	assert.NoError(t, err)
	assert.Equal(t, thingTD1, td2)
	// Update
	err = store.Replace(thingID, thingTD2)
	assert.NoError(t, err)
	td2, err = store.Get(thingID)
	assert.NoError(t, err)
	assert.Equal(t, thingTD2, td2)

	time.Sleep(time.Second * 10)

	// Delete
	store.Remove(thingID)
	_, err = store.Get(thingID)
	assert.Error(t, err)

	store.Close()
}
