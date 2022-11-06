package bucketstore_test

import (
	"fmt"
	"testing"

	"github.com/thanhpk/randstr"

	"github.com/hiveot/hub.go/pkg/logging"
)

// table with data size to run the benchmark with
var DataSizeTable = []struct {
	dataSize int
	textSize int
}{
	{dataSize: 1000, textSize: 100},
	{dataSize: 10000, textSize: 100},
	{dataSize: 100000, textSize: 100},
	{dataSize: 1000000, textSize: 100},
	{dataSize: 10000000, textSize: 100},
}

func Benchmark_createWriteBucket(b *testing.B) {
	logging.SetLogging("warning", "")

	for _, v := range DataSizeTable {
		//setup
		//testText := randstr.String(v.textSize)
		store, _ := openNewStore(testBackendType, testBackendPath)
		addDocs(store, testBucketID, v.dataSize)

		// run the test write bucket test
		b.Run(fmt.Sprintf("CreateWriteBucket datasize=%d", v.dataSize), func(b *testing.B) {
			//key := fmt.Sprintf("keyID-%d", rand.Intn(999999))
			for n := 0; n < b.N; n++ {
				bucket := store.GetWriteBucket(testBucketID)
				//_, _ = bucket.Get(key)
				bucket.Close(true)
			}

		})

		store.Close()
	}
}

func Benchmark_createReadBucket(b *testing.B) {
	logging.SetLogging("warning", "")

	for _, v := range DataSizeTable {
		//setup
		//testText := randstr.String(v.textSize)
		store, _ := openNewStore(testBackendType, testBackendPath)
		addDocs(store, testBucketID, v.dataSize)

		// run the test write bucket test
		b.Run(fmt.Sprintf("GetReadBucket datasize=%d", v.dataSize), func(b *testing.B) {
			//key := fmt.Sprintf("keyID-%d", rand.Intn(999999))
			for n := 0; n < b.N; n++ {
				bucket := store.GetReadBucket(testBucketID)
				bucket.Close(false)
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
		addDocs(store, testBucketID, v.dataSize)

		// run the test write bucket test
		b.Run(fmt.Sprintf("Bucket.Get datasize=%d;textSize=%d", v.dataSize, v.textSize), func(b *testing.B) {
			key := fmt.Sprintf("addDocs-5000")
			for n := 0; n < b.N; n++ {
				bucket := store.GetReadBucket(testBucketID)
				bucket.Get(key)
				bucket.Close(true)
			}
		})
		store.Close()
	}
}

func Benchmark_bucketSet(b *testing.B) {
	logging.SetLogging("warning", "")

	for _, v := range DataSizeTable {
		//setup
		testText := randstr.String(v.textSize)
		store, _ := openNewStore(testBackendType, testBackendPath)
		addDocs(store, testBucketID, v.dataSize)

		// run the test write bucket test
		b.Run(fmt.Sprintf("Bucket.Set datasize=%d;textSize=%d", v.dataSize, v.textSize), func(b *testing.B) {
			key := fmt.Sprintf("addDocs-5000")
			//key := fmt.Sprintf("keyID-%d", rand.Intn(999999))
			for n := 0; n < b.N; n++ {
				bucket := store.GetWriteBucket(testBucketID)
				bucket.Set(key, []byte(testText))
				bucket.Close(true)
			}
		})
		store.Close()
	}
}

func Benchmark_seekReadBucket(b *testing.B) {
	logging.SetLogging("warning", "")

	for _, v := range DataSizeTable {
		//setup
		//testText := randstr.String(v.textSize)
		store, _ := openNewStore(testBackendType, testBackendPath)
		addDocs(store, testBucketID, v.dataSize)

		// run the test write bucket test
		b.Run(fmt.Sprintf("ReadBucket.Seek datasize=%d", v.dataSize), func(b *testing.B) {
			key := fmt.Sprintf("addDocs-5000")
			for n := 0; n < b.N; n++ {
				bucket := store.GetReadBucket(testBucketID)
				cursor := bucket.Cursor()
				cursor.Seek(key)
				bucket.Close(false)
			}

		})

		store.Close()
	}
}
