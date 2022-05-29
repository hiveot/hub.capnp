package unpwstore_test

import (
	"fmt"
	"github.com/wostzone/wost-go/pkg/logging"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/authn/pkg/unpwstore"
)

const unpwFileName = "testunpwstore.passwd"

var unpwFilePath string

var configFolder string

// TestMain for all authn tests, setup of default folders and filenames
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	configFolder = path.Join(homeFolder, "config")

	// Make sure ACL and password files exist
	unpwFilePath = path.Join(configFolder, unpwFileName)
	fp, _ := os.Create(unpwFilePath)
	fp.Close()

	res := m.Run()
	os.Exit(res)
}

func TestOpenClosePWFile(t *testing.T) {
	fp, _ := os.Create(unpwFilePath)
	fp.Close()
	unpwStore := unpwstore.NewPasswordFileStore(unpwFilePath, "TestOpenClosePWFile")
	err := unpwStore.Open()
	assert.NoError(t, err)
	time.Sleep(time.Second * 1)
	assert.NoError(t, err)
	unpwStore.Close()
}

func TestSetPasswordTwoStores(t *testing.T) {
	const user1 = "user1"
	const user2 = "user2"
	const hash1 = "hash1"
	const hash2 = "hash2"
	fp, _ := os.Create(unpwFilePath)
	fp.Close()
	// create 2 separate stores
	pwStore1 := unpwstore.NewPasswordFileStore(unpwFilePath, "TestSetPasswordTwoStores-store1")
	err := pwStore1.Open()
	assert.NoError(t, err)
	pwStore2 := unpwstore.NewPasswordFileStore(unpwFilePath, "TestSetPasswordTwoStores-Store2")
	err = pwStore2.Open()
	assert.NoError(t, err)

	// set hash in store 1, should appear in store 2
	err = pwStore1.SetPasswordHash(user1, hash1)
	assert.NoError(t, err)
	// wait for reload
	time.Sleep(time.Second * 1)
	// check mode of pw file
	info, err := os.Stat(unpwFilePath)
	assert.NoError(t, err)
	mode := info.Mode()
	assert.Equal(t, 0600, int(mode), "file mode not 0600")

	// read back
	// force reload. Don't want to wait
	pwStore2.Reload()
	hash := pwStore2.GetPasswordHash(user1)
	assert.Equal(t, hash1, hash)

	// do it again but in reverse
	logrus.Infof("- do it again in reverse -")
	err = pwStore2.SetPasswordHash(user2, hash2)
	assert.NoError(t, err)
	time.Sleep(time.Second * 2)
	hash = pwStore1.GetPasswordHash(user2)
	assert.Equal(t, hash2, hash)
	// time.Sleep(time.Second * 1)

	assert.NoError(t, err)
	time.Sleep(time.Second * 1)
	pwStore1.Close()
	pwStore2.Close()
}

func TestNoPasswordFile(t *testing.T) {
	pwFile := path.Join(configFolder, "missingpasswordfile")
	pwStore := unpwstore.NewPasswordFileStore(pwFile, "TestNoPasswordFile")
	err := pwStore.Open()
	assert.Error(t, err)

	pwStore.Close()
}

// Load test if one writer with second reader
func TestConcurrentReadWrite(t *testing.T) {
	var wg sync.WaitGroup
	var i int

	// start with empty file
	fp, _ := os.Create(unpwFilePath)
	fp.Close()

	// two stores in parallel
	pwStore1 := unpwstore.NewPasswordFileStore(unpwFilePath, "TestConcurrentReadWrite-store1 (writer)")
	err := pwStore1.Open()
	assert.NoError(t, err)
	pwStore2 := unpwstore.NewPasswordFileStore(unpwFilePath, "TestConcurrentReadWrite-store2 (reader)")
	err = pwStore2.Open()
	assert.NoError(t, err)

	wg.Add(1)
	go func() {
		for i = 0; i < 30; i++ {
			thingID := fmt.Sprintf("thing-%d", i)
			err2 := pwStore1.SetPasswordHash(thingID, "hash1")
			time.Sleep(time.Millisecond * 1)
			if err2 != nil {
				assert.NoError(t, err2)
			}
		}
		wg.Done()
	}()
	wg.Wait()
	// time to catch up the file watcher debouncing
	time.Sleep(time.Second * 2)

	// both stores should be fully up to date
	assert.Equal(t, i, pwStore1.Count())
	assert.Equal(t, i, pwStore2.Count())

	//
	pwStore1.Close()
	pwStore2.Close()
}

func TestWritePwToTempFail(t *testing.T) {
	pws := make(map[string]string)
	pwStore1 := unpwstore.NewPasswordFileStore(unpwFilePath, "TestWritePwToTempFail")
	err := pwStore1.Open()
	assert.NoError(t, err)
	_, err = unpwstore.WritePasswordsToTempFile("/badfolder", pws)
	assert.Error(t, err)
	pwStore1.Close()
}
