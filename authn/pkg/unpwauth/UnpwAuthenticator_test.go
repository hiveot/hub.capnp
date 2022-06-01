package unpwauth_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/wostzone/wost-go/pkg/logging"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/wostzone/hub/authn/pkg/unpwauth"
	"github.com/wostzone/hub/authn/pkg/unpwstore"
)

const unpwFileName = "test.passwd"

var unpwFilePath string
var tempFolder string

// TestMain for all authn tests, setup of default folders and filenames
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	tempFolder = path.Join(os.TempDir(), "wost-authn-test")
	os.MkdirAll(tempFolder, 0700)

	// Make sure password file exist
	unpwFilePath = path.Join(tempFolder, unpwFileName)
	fp, _ := os.Create(unpwFilePath)
	fp.Close()
	// creating these files takes a bit of time,
	time.Sleep(time.Second)

	res := m.Run()
	if res == 0 {
		os.RemoveAll(tempFolder)
	}
	os.Exit(res)
}

// Create authn handler with empty username/pw
func createEmptyTestAuthenticator() *unpwauth.UnpwAuthenticator {
	fp, _ := os.Create(unpwFilePath)
	fp.Close()
	unpwStore := unpwstore.NewPasswordFileStore(unpwFilePath, "createEmptyTestAuthHandler")
	ah := unpwauth.NewUnPwAuthenticator(unpwStore)
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
	ah := unpwauth.NewUnPwAuthenticator(unpwStore)

	// opening the password store should fail
	err := ah.Start()
	assert.Error(t, err)
	ah.Stop()

	//
	ah = unpwauth.NewUnPwAuthenticator(nil)
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
	_, err := unpwauth.CreatePasswordHash(password1, "Badalgo", 0)
	assert.Error(t, err)
	_, err = unpwauth.CreatePasswordHash("", "", 0)
	assert.Error(t, err)

	hash, err := unpwauth.CreatePasswordHash("user1", unpwauth.PWHASH_ARGON2id, 0)
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
	hash, err := unpwauth.CreatePasswordHash(password1, unpwauth.PWHASH_BCRYPT, 0)
	assert.NoError(t, err)
	match := ah.VerifyPasswordHash(hash, password1, unpwauth.PWHASH_BCRYPT)
	assert.True(t, match)
	ah.Stop()
}
