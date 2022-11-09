package state_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/pkg/state"
)

const DefaultClientID = "statebench_text"
const DefaultBucketID = "default"

// Add records to the state store
func addRecords(store state.IStateService, clientID, bucketID string, count int) {
	const batchSize = 50000
	ctx := context.Background()
	client, _ := store.CapClientBucket(ctx, clientID, bucketID)
	nrBatches := (count / batchSize) + 1

	// Don't exceed the max transaction size
	for iBatch := 0; iBatch < nrBatches; iBatch++ {

		docs := make(map[string][]byte)
		for i := 0; i < batchSize && count > 0; i++ {
			k := randstr.String(12)
			v := randstr.Bytes(100)
			docs[k] = v
			count--
		}
		client.SetMultiple(ctx, docs)
		//client.Commit(nil)
	}
	client.Release()
}

// Table with data size to run the benchmark with
//
// KVStore write performance:
// DB records     set 1       set1000x1   setMultiple/1  setMultiple/1000
//     1K        0.3 usec      250 usec     0.1 usec         53 usec
//   100K        0.3 usec      290 usec     0.1 usec         70 usec
//
// BoltDB write performance:
// DB records     set 1        set1000x1   setMultiple/1  setMultiple/1000
//     1K        4.5 msec       4700 msec    4.6 msec         10.2
//   100K        6.7 msec       6621 msec    6.5 msec         29.8
//
// Pebble write performance:
// DB records     set 1        set1000x1   setMultiple/1  setMultiple/1000
//     1K        2.5 usec      2650 usec     1.8 usec        1970 usec
//   100K        3.2 usec      3450 usec     2.2 usec        3110 usec
//     1M        9.1 usec      5880 usec     1.9 usec        7470 usec
//
// --- via capnp ---
//
// KVStore via Capnp write performance:
// DB records     set 1        set1000x1   setMultiple/1  setMultiple/1000
//     1K        0.12 msec      118 msec      0.14 msec      4.3 msec
//   100K        0.12 msec      123 msec      0.13 msec      4.3 msec
//
// BoltDB via Capnp write performance:
// DB records     set 1        set1000x1   setMultiple/1  setMultiple/1000
//     1K        5.1 msec      4927 msec      5.1 msec        16 msec
//   100K        7.1 msec      6952 msec      6.9 msec        36 msec
//
// Pebble via Capnp write performance:
// DB records     set 1        set1000x1   setMultiple/1  setMultiple/1000   get 1     get1000x1
//     1K        0.13 msec      136 msec      0.15 msec       6.6 msec      0.13 msec   136 msec
//   100K        0.13 msec      136 msec      0.15 msec       7.9 msec      0.14 msec   132 msec
//
var DataSizeTable = []struct {
	dataSize int
	nrSets   int
}{
	{dataSize: 1000, nrSets: 1},
	{dataSize: 100000, nrSets: 1},
	//{dataSize: 1000000, nrSets: 1},
	{dataSize: 1000, nrSets: 1000},
	{dataSize: 100000, nrSets: 1000},
	//{dataSize: 1000000, nrSets: 1000},
}

// Generate random test data used to set and set multiple
type TestEl struct {
	key string
	val []byte
}

var testData = func() []TestEl {
	count := 100000
	data := make([]TestEl, count)
	for i := 0; i < count; i++ {
		key := randstr.String(10) // 10 char string
		val := randstr.Bytes(100) // 100 byte data
		data[i] = TestEl{key: key, val: val}
	}
	return data
}()

// test performance of N random set state
func BenchmarkSetState(b *testing.B) {
	const clientID1 = "test-client1"
	const appID = DefaultBucketID

	for _, tbl := range DataSizeTable {

		// setup
		logging.SetLogging("warning", "")
		ctx := context.Background()
		store, stopFn, err := createStateService(testUseCapnp)
		require.NoError(b, err)
		addRecords(store, clientID1, appID, tbl.dataSize)
		clientState, err := store.CapClientBucket(ctx, clientID1, appID)

		b.Run(fmt.Sprintf("SetState. Datasize=%d, #sets=%d", tbl.dataSize, tbl.nrSets),
			func(b *testing.B) {
				for n := 0; n < b.N; n++ {

					// range iterator adds approx 0.4 usec per call for 1M dataset
					for i := 0; i < tbl.nrSets; i++ {
						j := rand.Intn(100000)
						td := testData[j]
						err = clientState.Set(ctx, td.key, td.val)
						assert.NoError(b, err)
					}
				}
			})
		clientState.Release()
		err = stopFn()
		assert.NoError(b, err)
	}
}

// test performance of N random set state
func BenchmarkSetMultiple(b *testing.B) {
	const clientID1 = "test-client1"
	const appID = DefaultBucketID

	for _, tbl := range DataSizeTable {
		// setup
		logging.SetLogging("warning", "")
		ctx := context.Background()
		store, stopFn, err := createStateService(testUseCapnp)
		require.NoError(b, err)
		// prepopulate the store with records
		addRecords(store, clientID1, appID, tbl.dataSize)

		// build a set of data to test with
		multiple := make(map[string][]byte)
		_ = multiple
		for i := 0; i < tbl.nrSets; i++ {
			td := testData[i]
			multiple[td.key] = td.val
		}

		clientState, err := store.CapClientBucket(ctx, clientID1, appID)

		b.Run(fmt.Sprintf("SetMultiple. Datasize=%d, #sets=%d", tbl.dataSize, tbl.nrSets),
			func(b *testing.B) {
				// test set
				for n := 0; n < b.N; n++ {
					err = clientState.SetMultiple(ctx, multiple)
					assert.NoError(b, err)
				}
			})

		clientState.Release()
		err = stopFn()
		assert.NoError(b, err)
	}
}

// test performance of N single get/sets
func BenchmarkGetState(b *testing.B) {
	const clientID1 = "test-client1"
	const appID = "test-app"
	const key1 = "key1"
	var val1 = []byte("value 1")

	for _, v := range DataSizeTable {

		// setup
		logging.SetLogging("warning", "")
		ctx := context.Background()
		store, stopFn, err := createStateService(testUseCapnp)
		require.NoError(b, err)
		addRecords(store, clientID1, appID, v.dataSize)
		clientState, err := store.CapClientBucket(ctx, clientID1, appID)
		clientState.Set(ctx, key1, val1)

		// create the client, update and close
		b.Run(fmt.Sprintf("GetState. Datasize=%d, %d gets", v.dataSize, v.nrSets), func(b *testing.B) {
			// test get
			for n := 0; n < b.N; n++ {
				assert.NoError(b, err)
				for i := 0; i < v.nrSets; i++ {
					val2, err2 := clientState.Get(ctx, key1)
					assert.NoError(b, err2)
					assert.Equal(b, val1, val2)
				}
			}
		})
		clientState.Release()

		err = stopFn()
		assert.NoError(b, err)
	}
}
