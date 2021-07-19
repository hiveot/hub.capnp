package auth_test

import (
	"fmt"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/pkg/auth"
)

const pwFileName = "pw-test.conf"

func TestOpenClosePWFile(t *testing.T) {
	pwFile := path.Join(configFolder, pwFileName)
	fp, _ := os.Create(pwFile)
	fp.Close()
	pwStore := auth.NewPasswordFileStore(pwFile)
	err := pwStore.Open()
	assert.NoError(t, err)
	time.Sleep(time.Second * 1)
	assert.NoError(t, err)
	pwStore.Close()
}

func TestSetPasswordTwoStores(t *testing.T) {
	const user1 = "user1"
	const user2 = "user2"
	const hash1 = "hash1"
	const hash2 = "hash2"
	pwFile := path.Join(configFolder, pwFileName)
	fp, _ := os.Create(pwFile)
	fp.Close()
	// create 2 separate stores
	pwStore1 := auth.NewPasswordFileStore(pwFile)
	err := pwStore1.Open()
	assert.NoError(t, err)
	pwStore2 := auth.NewPasswordFileStore(pwFile)
	err = pwStore2.Open()
	assert.NoError(t, err)

	// set hash in store 1, should appear in store 2
	err = pwStore1.SetPasswordHash(user1, hash1)
	assert.NoError(t, err)
	// wait for reload
	time.Sleep(time.Second * 1)
	// check mode of pw file
	info, err := os.Stat(pwFile)
	assert.NoError(t, err)
	mode := info.Mode()
	assert.Equal(t, 0600, int(mode), "file mode not 0600")

	// read back
	hash := pwStore2.GetPasswordHash(user1)
	assert.Equal(t, hash1, hash)

	// do it again but in reverse
	logrus.Infof("- do it again in reverse -")
	err = pwStore2.SetPasswordHash(user2, hash2)
	assert.NoError(t, err)
	time.Sleep(time.Second * 1)
	hash = pwStore1.GetPasswordHash(user2)
	assert.Equal(t, hash2, hash)
	time.Sleep(time.Second * 1)

	assert.NoError(t, err)
	pwStore1.Close()
	pwStore2.Close()
}

func TestNoPasswordFile(t *testing.T) {
	pwFile := path.Join(configFolder, "missingpasswordfile")
	pwStore := auth.NewPasswordFileStore(pwFile)
	err := pwStore.Open()
	assert.Error(t, err)

	pwStore.Close()
}

// Load test if one writer with second reader
func TestConcurrentReadWrite(t *testing.T) {
	var wg sync.WaitGroup
	var i int

	// start with empty file
	pwFile := path.Join(configFolder, pwFileName)
	fp, _ := os.Create(pwFile)
	fp.Close()

	// two stores in parallel
	pwStore1 := auth.NewPasswordFileStore(pwFile)
	err := pwStore1.Open()
	assert.NoError(t, err)
	pwStore2 := auth.NewPasswordFileStore(pwFile)
	err = pwStore2.Open()
	assert.NoError(t, err)

	wg.Add(1)
	go func() {
		for i = 0; i < 1000; i++ {
			thingID := fmt.Sprintf("thing-%d", i)
			err2 := pwStore1.SetPasswordHash(thingID, "hash1")
			time.Sleep(time.Millisecond)
			if err2 != nil {
				assert.NoError(t, err2)
			}
		}
		wg.Done()
	}()
	wg.Wait()
	// time to catch up the file watcher debouncing
	time.Sleep(time.Second)

	// both stores should be fully up to date
	assert.Equal(t, i, pwStore1.Count())
	assert.Equal(t, i, pwStore2.Count())

	//
	pwStore1.Close()
	pwStore2.Close()
}

func TestWritePwToTempFail(t *testing.T) {

	pwFile := path.Join(configFolder, pwFileName)
	pwStore1 := auth.NewPasswordFileStore(pwFile)
	err := pwStore1.Open()
	assert.NoError(t, err)
	_, err = pwStore1.WriteToTemp("/badfolder")
	assert.Error(t, err)
	aclStore.Close()
}
