package clientconfigstore_test

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wostzone/hub/authn/pkg/clientconfigstore"
	"github.com/wostzone/wost-go/pkg/logging"
)

var tempFolder string

// TestMain determines the store location
// Used for all test cases in this package
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	tempFolder = path.Join(os.TempDir(), "wost-authn-test")
	os.MkdirAll(tempFolder, 0700)

	res := m.Run()
	if res == 0 {
		os.RemoveAll(tempFolder)
	}
	os.Exit(res)
}

func TestOpenClose(t *testing.T) {
	cfgStore := clientconfigstore.NewClientConfigStore(tempFolder)
	err := cfgStore.Open()
	assert.NoError(t, err)
	cfgStore.Close()
}

func TestWriteConfig(t *testing.T) {
	const user1ID = "user1"
	const app1ID = "app1"
	const configData = "Hello world"
	cfgStore := clientconfigstore.NewClientConfigStore(tempFolder)
	err := cfgStore.Open()
	assert.NoError(t, err)

	err = cfgStore.Put(user1ID, app1ID, configData)
	assert.NoError(t, err)

	rxData := cfgStore.Get(user1ID, app1ID)
	assert.Equal(t, configData, rxData)
	cfgStore.Close()
}

func TestReadMissingConfig(t *testing.T) {
	const user1ID = "user1"
	const app2ID = "app2"
	cfgStore := clientconfigstore.NewClientConfigStore(tempFolder)
	err := cfgStore.Open()
	assert.NoError(t, err)

	rxData := cfgStore.Get(user1ID, app2ID)
	assert.Empty(t, rxData)
	cfgStore.Close()
}
