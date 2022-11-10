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

// Benchmark with effects of database size
//
//              DB size   records        kv (us)       pebble (us)   bbolt (us)
// SetState        1K,       1             0.6            4.0             4900
//               100K,       1             0.7            4.2             7000
//                 1M,       1                            9.1
//                 1K      1000          600           4000            4800000 (4.8 sec!)
//               100K      1000          700           5300            7000000 (7 sec!)
//                 1M      1000                        5900
// SetMultiple     1K,       1             0.1            1.8             4700
//               100K,       1             0.2            2.2             6600
//                 1M,       1                            1.9
//                 1K      1000          170           1900              11000
//               100K      1000          330           2800              31000
//                 1M,     1000                        7470
// GetState        1K,       1             0.5            0.9                1.6
//               100K,       1             0.6            0.9                1.7
//                 1K      1000          530            870               1500
//               100K      1000          590            920               1700
// GetMultiple     1K,       1
//               100K,       1
//                 1K      1000
//               100K      1000

// Benchmark with use of capnp. Note timing in msec
//
//              DB size   records        kv (ms)       pebble (ms)     bbolt (ms)
// SetState        1K,       1             0.1            0.1              5.0
//               100K,       1             0.1            0.1              7.1
//                 1K      1000          120            140             4900   (4.9 sec!)
//               100K      1000          120            140             7000   (7 sec!)
// SetMultiple     1K,       1             0.14           0.15             5.1
//               100K,       1             0.13           0.15             6.9
//                 1K      1000            4.3            6.6             16
//               100K      1000            4.3            7.9             36
// GetState        1K,       1             0.13           0.13             0.13
//               100K,       1             0.13           0.13             0.13
//                 1K      1000          130            130              130
//               100K      1000          130            130              130
// GetMultiple     1K,       1
//               100K,       1
//                 1K      1000
//               100K      1000
//
// Observations:
//  - transaction write of bbolt is very costly. Use setmultiple or performance will be insufficient
//  - the capnp RPC over Unix Domain Sockets call overhead is around 0.13 msec.

var DataSizeTable = []struct {
	dataSize int
	nrSets   int
}{
	{dataSize: 1000, nrSets: 1},
	{dataSize: 100000, nrSets: 1},
	{dataSize: 1000, nrSets: 1000},
	{dataSize: 100000, nrSets: 1000},
	//{dataSize: 1000000, nrSets: 1},
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
