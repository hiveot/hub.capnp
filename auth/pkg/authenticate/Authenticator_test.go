package authenticate_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/auth/pkg/authenticate"
	"github.com/wostzone/hub/auth/pkg/unpwstore"
	"github.com/wostzone/hub/lib/client/pkg/config"
)

const unpwFileName = "test.passwd"

var unpwFilePath string

// TestMain for all auth tests, setup of default folders and filenames
func TestMain(m *testing.M) {
	config.SetLogging("info", "")
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	configFolder := path.Join(homeFolder, "config")

	// Make sure password file exist
	unpwFilePath = path.Join(configFolder, unpwFileName)
	fp, _ := os.Create(unpwFilePath)
	fp.Close()
	// creating these files takes a bit of time,
	time.Sleep(time.Second)

	res := m.Run()
	os.Exit(res)
}

// Create auth handler with empty username/pw
func createEmptyTestAuthenticator() *authenticate.Authenticator {
	fp, _ := os.Create(unpwFilePath)
	fp.Close()
	unpwStore := unpwstore.NewPasswordFileStore(unpwFilePath, "createEmptyTestAuthHandler")
	ah := authenticate.NewAuthenticator(unpwStore)
	return ah
}

func TestAuthenticatorStartStop(t *testing.T) {
	logrus.Infof("---TestAuthenticatorStartStop---")
	ah := createEmptyTestAuthenticator()
	err := ah.Start()
	time.Sleep(time.Second * 1)
	assert.NoError(t, err)
	ah.Stop()
}

func TestAuthHandlerBadStart(t *testing.T) {
	logrus.Infof("---TestAuthHandlerBadStart---")
	unpwStore := unpwstore.NewPasswordFileStore("/bad/file", "TestAuthHandlerBadStart")
	ah := authenticate.NewAuthenticator(unpwStore)

	// opening the password store should fail
	err := ah.Start()
	assert.Error(t, err)
	ah.Stop()

	//
	ah = authenticate.NewAuthenticator(nil)
	err = ah.Start()
	assert.Error(t, err)

	// should not blow up without password store
	match := ah.VerifyUsernamePassword("user1", "user1")
	assert.False(t, match)

}

func TestUnpwMatch(t *testing.T) {
	logrus.Infof("---TestUnpwMatch---")
	userName1 := "user1" // as in test file
	userName2 := "user2" // as in test file
	password1 := "password1"
	password2 := "password2"

	ah := createEmptyTestAuthenticator()
	ah.Start()
	err := ah.SetPassword(userName1, password1)
	assert.NoError(t, err)

	err = ah.SetPassword(userName2, password2)
	assert.NoError(t, err)

	err = ah.SetPassword(userName2, "")
	assert.Error(t, err)

	err = ah.SetPassword("", password2)
	assert.Error(t, err)

	match := ah.VerifyUsernamePassword(userName1, password1)
	assert.True(t, match)

	match = ah.VerifyUsernamePassword(userName1, password2)
	assert.False(t, match)

	match = ah.VerifyUsernamePassword("notauser", password1)
	assert.False(t, match)

	ah.Stop()
}

func TestBadHashAlgo(t *testing.T) {
	logrus.Infof("---TestBadHashAlgo---")
	password1 := "password1"
	_, err := authenticate.CreatePasswordHash(password1, "Badalgo", 0)
	assert.Error(t, err)
	_, err = authenticate.CreatePasswordHash("", "", 0)
	assert.Error(t, err)

	hash, err := authenticate.CreatePasswordHash("user1", authenticate.PWHASH_ARGON2id, 0)
	assert.NoError(t, err)
	ah := createEmptyTestAuthenticator()
	match := ah.VerifyPasswordHash(hash, "password1", "badalgo")
	assert.False(t, match)
}

func TestBCrypt(t *testing.T) {
	logrus.Infof("---TestBCrypt---")
	var password1 = "password1"
	ah := createEmptyTestAuthenticator()
	err := ah.Start()
	assert.NoError(t, err)
	hash, err := authenticate.CreatePasswordHash(password1, authenticate.PWHASH_BCRYPT, 0)
	assert.NoError(t, err)
	match := ah.VerifyPasswordHash(hash, password1, authenticate.PWHASH_BCRYPT)
	assert.True(t, match)
	ah.Stop()
}
