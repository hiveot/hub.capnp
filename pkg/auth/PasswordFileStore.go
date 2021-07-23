package auth

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/watcher"
)

// The default filename of the username/password file store
const DefaultUnpwFilename = "unpw.passwd"

// PasswordFileStore stores a list of user login names and their password info
// It includes a file watcher to automatically reload on update.
type PasswordFileStore struct {
	passwords map[string]string // loginName: bcrypted(password)
	storePath string
	watcher   *fsnotify.Watcher
	mutex     sync.RWMutex
}

// Close the store
func (pwStore *PasswordFileStore) Close() {
	if pwStore.watcher != nil {
		pwStore.watcher.Close()
		pwStore.watcher = nil
	}
}

// Count nr of entries in the store
func (pwStore *PasswordFileStore) Count() int {
	return len(pwStore.passwords)
}

// Return the stored bcrypt encoded password hash for the given user
// returns hash or "" if user not found
func (pwStore *PasswordFileStore) GetPasswordHash(username string) string {
	hash := pwStore.passwords[username]
	return hash
}

// Open the store
// This reads the acl file and subscribes to file changes
func (pwStore *PasswordFileStore) Open() error {
	err := pwStore.Reload()
	if err == nil {
		pwStore.watcher, err = watcher.WatchFile(pwStore.storePath, pwStore.Reload)
	}
	return err
}

// Reload the password store from file and subscribe to file changes
// If subscription already exists it is removed and renewed.
//  File format:  <loginname>:bcrypt(passwd)
// Returns error if the file could not be opened
func (pwStore *PasswordFileStore) Reload() error {
	pwStore.mutex.RLock()
	defer pwStore.mutex.RUnlock()
	logrus.Infof("PasswordFileStore.Reload: Reloading passwords from %s", pwStore.storePath)

	pwList := make(map[string]string)
	file, err := os.Open(pwStore.storePath)
	if err != nil {
		err := fmt.Errorf("Reload: Failed to open password file: %s", err)
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
	logrus.Infof("Reload: loaded %d passwords", len(pwList))
	pwStore.passwords = pwList

	// // Renew the watcher after because a file rename assigns a new inode.
	// if pwStore.watcher != nil {
	// 	pwStore.watcher.Close()
	// }
	// pwStore.watcher, err = watcher.WatchFile(pwStore.storePath, pwStore.Reload)

	return err
}

// Add/update the password hash for the given login ID
// Intended for use by administrators to add a new user or clients to update their password
func (pwStore *PasswordFileStore) SetPasswordHash(loginID string, hash string) error {
	pwStore.mutex.Lock()
	defer pwStore.mutex.Unlock()
	if pwStore.passwords == nil {
		logrus.Panic("Use of password store before open")
	}

	pwStore.passwords[loginID] = hash
	folder := path.Dir(pwStore.storePath)
	tmpPath, err := pwStore.WriteToTemp(folder)
	if err != nil {
		logrus.Infof("SetPasswordHash write: %s", err)
		return err
	}
	logrus.Infof("rename temp pwfile: (loginID=%s): %s", loginID, tmpPath)

	err = os.Rename(tmpPath, pwStore.storePath)
	if err != nil {
		logrus.Infof("SetPasswordHash rename: %s", err)
		return err
	}

	logrus.Infof("PasswordFileStore.SetPasswordHash: password hash for loginID '%s' updated", loginID)
	return err
}

// Write the ACL store to temp file in the password folder
// This returns the name of the new temp file
func (pwStore *PasswordFileStore) WriteToTemp(folder string) (tempFileName string, err error) {
	file, err := os.CreateTemp(folder, "hub-pwfilestore")
	// file, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		err := fmt.Errorf("PasswordFileStore.Write: Failed open temp password file: %s", err)
		logrus.Error(err)
		return "", err
	}
	tempFileName = file.Name()

	defer file.Close()
	writer := bufio.NewWriter(file)
	for key, value := range pwStore.passwords {
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

// New instance of a file based ACL store
// Note: this store is intended for one writer and many readers.
// Multiple concurrent writes are not supported and might lead to one write being ignored.
//  filepath location of the file store. See also DefaultUnpwFilename for the recommended name
func NewPasswordFileStore(filepath string) *PasswordFileStore {
	store := &PasswordFileStore{
		storePath: filepath,
	}
	return store
}
