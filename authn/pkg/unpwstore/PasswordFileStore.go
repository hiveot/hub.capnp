package unpwstore

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/serve/pkg/watcher"
)

// DefaultPasswordFile is the recommended password filename for Hub authentication
const DefaultPasswordFile = "hub.passwd"

// PasswordFileStore stores a list of user login names and their password info
// It includes a file watcher to automatically reload on update.
type PasswordFileStore struct {
	clientID  string            // the user of the store for logging
	passwords map[string]string // loginName: bcrypted(password)
	storePath string
	watcher   *fsnotify.Watcher
	mutex     sync.RWMutex
}

// Close the store
func (pwStore *PasswordFileStore) Close() {
	logrus.Infof("PasswordFileStore.Close. clientID='%s'", pwStore.clientID)
	if pwStore.watcher != nil {
		pwStore.watcher.Close()
		pwStore.watcher = nil
	}
}

// Count nr of entries in the store
func (pwStore *PasswordFileStore) Count() int {
	pwStore.mutex.RLock()
	defer pwStore.mutex.RUnlock()

	return len(pwStore.passwords)
}

// GetPasswordHash returns the stored encoded password hash for the given user
// returns hash or "" if user not found
func (pwStore *PasswordFileStore) GetPasswordHash(username string) string {
	pwStore.mutex.RLock()
	defer pwStore.mutex.RUnlock()
	hash := pwStore.passwords[username]
	return hash
}

// Open the store
// This reads the password file and subscribes to file changes
func (pwStore *PasswordFileStore) Open() error {
	logrus.Infof("PasswordFileStore.Open. clientID='%s'", pwStore.clientID)
	err := pwStore.Reload()
	if err == nil {
		pwStore.watcher, err = watcher.WatchFile(pwStore.storePath, pwStore.Reload, pwStore.clientID)
	}
	return err
}

// Reload the password store from file and subscribe to file changes
// If subscription already exists it is removed and renewed.
//  File format:  <loginname>:bcrypt(passwd)
// Returns error if the file could not be opened
func (pwStore *PasswordFileStore) Reload() error {
	logrus.Infof("PasswordFileStore.Reload: clientID='%s', Reloading passwords from '%s'",
		pwStore.clientID, pwStore.storePath)
	pwStore.mutex.Lock()
	defer pwStore.mutex.Unlock()

	pwList := make(map[string]string)
	file, err := os.Open(pwStore.storePath)
	if err != nil {
		err := fmt.Errorf("PasswordFileStore.Reload: clientID='%s', Failed to open password file: %s", pwStore.clientID, err)
		logrus.Error(err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			// ignore
		} else {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				username := parts[0]
				pass := parts[1]
				pwList[username] = pass
			}
		}
	}
	logrus.Infof("Reload: clientID='%s', loaded %d passwords", pwStore.clientID, len(pwList))

	pwStore.passwords = pwList

	return err
}

// SetPasswordHash adds/updates the password hash for the given login ID
// Intended for use by administrators to add a new user or clients to update their password
func (pwStore *PasswordFileStore) SetPasswordHash(loginID string, hash string) error {
	logrus.Infof("PasswordFileStore.SetPasswordHash clientID='%s', for login '%s'", pwStore.clientID, loginID)
	pwStore.mutex.Lock()
	defer pwStore.mutex.Unlock()
	if pwStore.passwords == nil {
		logrus.Panic("Use of password store before open")
	}
	pwStore.passwords[loginID] = hash

	folder := path.Dir(pwStore.storePath)
	tmpPath, err := WritePasswordsToTempFile(folder, pwStore.passwords)
	if err != nil {
		logrus.Errorf("SetPasswordHash clientID='%s'. loginID='%s' write to temp failed: %s",
			pwStore.clientID, loginID, err)
		return err
	}

	err = os.Rename(tmpPath, pwStore.storePath)
	if err != nil {
		logrus.Errorf("SetPasswordHash clientID='%s'. loginID='%s' rename to password file failed: %s",
			pwStore.clientID, loginID, err)
		return err
	}

	logrus.Infof("PasswordFileStore.SetPasswordHash: clientID='%s'. password hash for loginID '%s' updated", pwStore.clientID, loginID)
	return err
}

// WritePasswordsToTempFile write the given passwords to temp file in the given folder
// This returns the name of the new temp file.
func WritePasswordsToTempFile(
	folder string, passwords map[string]string) (tempFileName string, err error) {

	file, err := os.CreateTemp(folder, "hub-pwfilestore")

	// file, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		err := fmt.Errorf("PasswordFileStore.Write: Failed open temp password file: %s", err)
		logrus.Error(err)
		return "", err
	}
	tempFileName = file.Name()
	logrus.Infof("PasswordFileStore.WritePasswordsToTempFile: %s", tempFileName)

	defer file.Close()
	writer := bufio.NewWriter(file)

	for key, value := range passwords {
		_, err = writer.WriteString(fmt.Sprintf("%s:%s\n", key, value))
		if err != nil {
			err := fmt.Errorf("PasswordFileStore.Write: Failed writing password file: %s", err)
			logrus.Error(err)
			return tempFileName, err
		}
	}
	writer.Flush()
	return tempFileName, err
}

// NewPasswordFileStore creates a new instance of a file based username/password store
// Note: this store is intended for one writer and many readers.
// Multiple concurrent writes are not supported and might lead to one write being ignored.
//  filepath location of the file store. See also DefaultPasswordFile for the recommended name
//  clientID is authservice ID to include in logging
func NewPasswordFileStore(filepath string, clientID string) *PasswordFileStore {
	store := &PasswordFileStore{
		clientID:  clientID,
		storePath: filepath,
	}
	return store
}
