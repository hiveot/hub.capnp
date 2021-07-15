package auth

import (
	"io/ioutil"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// A group is a map of clients and roles
type AclGroup map[string]string

// AclStoreFile stores ACL list in file.
// It includes a file watcher to automatically reload on update.
type AclStoreFile struct {
	Groups       map[string]*AclGroup `yaml:"groups"`
	storePath    string
	watcher      *fsnotify.Watcher
	clientGroups map[string](map[string]string) // list of groups the client is a member of. Updated on load.
}

// Close the store
func (aclStore *AclStoreFile) Close() {
	if aclStore.watcher != nil {
		aclStore.watcher.Close()
		aclStore.watcher = nil
	}
}

// GetGroups returns a list of groups a thing or user is a member of
func (aclStore *AclStoreFile) GetGroups(clientID string) []string {
	groupsMemberOf := []string{}

	cg := aclStore.clientGroups[clientID]
	for groupName, _ := range cg {
		groupsMemberOf = append(groupsMemberOf, groupName)
	}
	return groupsMemberOf
}

// Get highest role of a user has in a list of group
// Intended to get client permissions in case of overlapping groups
func (aclStore *AclStoreFile) GetRole(clientID string, groupIDs []string) string {
	highestRole := GroupRoleNone

	cg := aclStore.clientGroups[clientID]

	for _, groupID := range groupIDs {
		role := cg[groupID]
		if isRoleGreaterEqual(role, highestRole) {
			highestRole = role
		}
	}
	return highestRole
}

// isRoleGreaterEqual returns true if a user role has same or greater permissions
// than the minimum role.
func isRoleGreaterEqual(role string, minRole string) bool {
	if minRole == GroupRoleNone || role == minRole {
		return true
	}
	if minRole == GroupRoleViewer && role != GroupRoleNone {
		return true
	}
	if minRole == GroupRoleEditor && (role == GroupRoleManager) {
		return true
	}
	return false
}

// Open the store
// This reads the acl file and subscribes to file changes
func (aclStore *AclStoreFile) Open() error {
	aclStore.Reload()
	watcher, err := aclStore.Watch(aclStore.storePath, aclStore.Reload)
	aclStore.watcher = watcher

	return err
}

// reload the ACL store from file
func (aclStore *AclStoreFile) Reload() error {
	// TODO: lock the file before reading
	raw, err := ioutil.ReadFile(aclStore.storePath)
	if err != nil {
		logrus.Errorf("AclStoreFile.Reload '%s': Error opening the ACL file: %s", aclStore.storePath, err)
		return err
	}
	err = yaml.Unmarshal(raw, aclStore)
	if err != nil {
		logrus.Errorf("AclStoreFile.Reload '%s': Error parsing the ACL file: %s", aclStore.storePath, err)
		return err
	}
	logrus.Infof("AclStoreFile.Reload '%s': Success", aclStore.storePath)

	// update index to lookup the role of a client in a group:
	// eg: [clientID] = {group:role, group:role, ...}
	clientGroups := make(map[string](map[string]string))
	for groupName, group := range aclStore.Groups {
		for client, role := range *group {
			cg := clientGroups[client]
			if cg == nil {
				cg = make(map[string]string)
				clientGroups[client] = cg
			}
			cg[groupName] = role
		}
	}
	aclStore.clientGroups = clientGroups
	return nil
}

// Watch the path for changes and invoke the callback.
// This debounces multiple successive changes to a file before invoking the callback
//  path to watch
//  handler to invoke on change
// This returns the watcher. Close it when done.
func (aclStore *AclStoreFile) Watch(path string, handler func() error) (*fsnotify.Watcher, error) {
	watcher, _ := fsnotify.NewWatcher()
	// The callback timer debounces multiple changes to the config file
	callbackTimer := time.AfterFunc(0, func() {
		logrus.Debug("AuthStoreFile.Watch: invoking callback")
		handler()
	})
	callbackTimer.Stop() // don't start yet

	err := watcher.Add(aclStore.storePath)
	if err != nil {
		logrus.Errorf("AuthStoreFile.Watch: unable to watch for changes: %s", err)
		return watcher, err
	}
	// defer watcher.Close()

	// done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// don't really care what the change it, after 100msec the file will reload
				logrus.Debugf("Watch: event: %s. Modified file: %s", event, event.Name)
				callbackTimer.Reset(time.Millisecond * 100)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Errorf("Watch: Error: %s", err)
			}
		}
	}()
	err = watcher.Add(path)
	if err != nil {
		logrus.Errorf("Watch: error %s", err)
	}
	// <-done
	return watcher, nil
}

// New instance of a file based ACL store
func NewAclStoreFile(filepath string) IAclStoreReader {
	store := &AclStoreFile{
		storePath: filepath,
	}
	return store
}
