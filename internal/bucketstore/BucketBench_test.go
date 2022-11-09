package bucketstore_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"

	"github.com/hiveot/hub.go/pkg/logging"
)

// KVStore performance:
// DB records     set 1       set1000x1   setMultiple/1  setMultiple/1000 seek
//     1K        0.2 usec       66 usec
//   100K        0.2 usec       91 usec
//
// BoltDB performance:
// DB records     set 1        set1000x1   setMultiple/1  setMultiple/1000
//     1K        5.2 msec       5130 msec
//   100K       11.7 msec      11700 msec
//
// Pebble performance:
// DB records     set 1        set1000x1   setMultiple/1  setMultiple/1000
//     1K        2.0 usec      2228 usec
//   100K        1.9 usec      2610 usec
//

// table with data size to run the benchmark with
var DataSizeTable = []struct {
	dataSize int
	textSize int
	nrSteps  int
}{
	{dataSize: 1000, textSize: 100, nrSteps: 1},
	{dataSize: 1000, textSize: 100, nrSteps: 1000},
	{dataSize: 100000, textSize: 100, nrSteps: 1},
	{dataSize: 100000, textSize: 100, nrSteps: 1000},
	//{dataSize: 1000000, textSize: 100},
	//{dataSize: 10000000, textSize: 100},
}

// Generate random test data used to set and set multiple
type TestEl struct {
	key string
	val []byte
}

var testData = func() []TestEl {
	count := 1000000
	data := make([]TestEl, count)
	for i := 0; i < count; i++ {
		key := randstr.String(10) // 10 char string
		val := randstr.Bytes(100) // 100 byte data
		data[i] = TestEl{key: key, val: val}
	}
	return data
}()

func Benchmark_createBucket(b *testing.B) {
	logging.SetLogging("warning", "")

	for _, v := range DataSizeTable {
		//setup
		//testText := randstr.String(v.textSize)
		store, _ := openNewStore(testBackendType, testBackendPath)
		addDocs(store, testBucketID, v.dataSize)

		// run the test write bucket test
		b.Run(fmt.Sprintf("CreateBucket datasize=%d", v.dataSize), func(b *testing.B) {
			//key := fmt.Sprintf("keyID-%d", rand.Intn(999999))
			for n := 0; n < b.N; n++ {
				bucket := store.GetBucket(testBucketID)
				//_, _ = bucket.Get(key)
				bucket.Close()
			}
		})
		store.Close()
	}
}

func Benchmark_BucketGet(b *testing.B) {
	logging.SetLogging("warning", "")

	for _, v := range DataSizeTable {
		//setup
		store, _ := openNewStore(testBackendType, testBackendPath)
		err := addDocs(store, testBucketID, v.dataSize)
		assert.NoError(b, err)

		// run the test write bucket test
		b.Run(fmt.Sprintf("Bucket.Get datasize=%d;textSize=%d", v.dataSize, v.textSize), func(b *testing.B) {
			key := fmt.Sprintf("addDocs-5000")
			for n := 0; n < b.N; n++ {
				bucket := store.GetBucket(testBucketID)
				for i := 0; i < v.nrSteps; i++ {
					_, err := bucket.Get(key)
					assert.NoError(b, err)
				}
				err = bucket.Close()
				assert.NoError(b, err)
			}
		})
		err = store.Close()
		assert.NoError(b, err)
	}
}

func Benchmark_bucketSet(b *testing.B) {
	logging.SetLogging("warning", "")

	for _, v := range DataSizeTable {
		//setup
		//testText := randstr.String(v.textSize)
		store, _ := openNewStore(testBackendType, testBackendPath)
		err := addDocs(store, testBucketID, v.dataSize)
		assert.NoError(b, err)

		// run the test write bucket test
		b.Run(fmt.Sprintf("Bucket.Set datasize=%d;nrSteps=%d",
			v.dataSize, v.nrSteps), func(b *testing.B) {
			for n := 0; n < b.N; n++ {

				bucket := store.GetBucket(testBucketID)

				for i := 0; i < v.nrSteps; i++ {
					td := testData[i]
					err = bucket.Set(td.key, td.val)
					assert.NoError(b, err)
				}
				err = bucket.Close()
				assert.NoError(b, err)
			}
		})
		err = store.Close()
		assert.NoError(b, err)
	}
}

func Benchmark_seekReadBucket(b *testing.B) {
	logging.SetLogging("warning", "")

	for _, v := range DataSizeTable {
		//setup
		//testText := randstr.String(v.textSize)
		store, _ := openNewStore(testBackendType, testBackendPath)
		err := addDocs(store, testBucketID, v.dataSize)
		assert.NoError(b, err)

		// run the test write bucket test
		b.Run(fmt.Sprintf("ReadBucket.Seek datasize=%d,steps=%d", v.dataSize, v.nrSteps), func(b *testing.B) {
			key := fmt.Sprintf("addDocs-5000")
			for n := 0; n < b.N; n++ {
				bucket := store.GetBucket(testBucketID)
				cursor, err := bucket.Cursor()
				require.NoError(b, err)

				for i := 0; i < v.nrSteps; i++ {
					key2, val2 := cursor.Seek(key)
					_ = key2
					_ = val2
					assert.NoError(b, err)
				}

				cursor.Release()
				err = bucket.Close()
				assert.NoError(b, err)
			}

		})

		err = store.Close()
		assert.NoError(b, err)
	}
}
