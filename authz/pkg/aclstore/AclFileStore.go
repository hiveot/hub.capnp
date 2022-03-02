package aclstore

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/authz/pkg/authorize"
	"github.com/wostzone/hub/lib/serve/pkg/watcher"
	"gopkg.in/yaml.v3"
)

// DefaultAclFile recommended ACL filename for Hub authentication
const DefaultAclFile = "hub.acl"

// AclGroup is a map of group clients and roles
type AclGroup map[string]string // map of clientID:role

// AclFileStore stores ACL list in file.
// It includes a file watcher to automatically reload on update.
type AclFileStore struct {
	clientID     string              // for logging
	Groups       map[string]AclGroup `yaml:"groups"` // store by group ID
	storePath    string
	watcher      *fsnotify.Watcher
	mutex        sync.RWMutex
	clientGroups map[string]map[string]string // list of groups the client is a member of. Updated on load.
}

// Close the store
func (aclStore *AclFileStore) Close() {
	logrus.Infof("AclFileStore.Close: clientID='%s'", aclStore.clientID)
	aclStore.mutex.Lock()
	defer aclStore.mutex.Unlock()
	if aclStore.watcher != nil {
		aclStore.watcher.Close()
		aclStore.watcher = nil
	}
}

// GetGroups returns a list of groups a thing or user is a member of
func (aclStore *AclFileStore) GetGroups(clientID string) []string {
	groupsMemberOf := []string{}

	aclStore.mutex.RLock()
	defer aclStore.mutex.RUnlock()
	cg := aclStore.clientGroups[clientID]
	for groupName := range cg {
		groupsMemberOf = append(groupsMemberOf, groupName)
	}
	return groupsMemberOf
}

// GetRole returns the highest role of a user has in a list of group
// Intended to get client permissions in case of overlapping groups
func (aclStore *AclFileStore) GetRole(clientID string, groupIDs []string) string {
	highestRole := authorize.GroupRoleNone

	aclStore.mutex.RLock()
	defer aclStore.mutex.RUnlock()

	cg := aclStore.clientGroups[clientID]

	for _, groupID := range groupIDs {
		role := cg[groupID]
		if IsRoleGreaterEqual(role, highestRole) {
			highestRole = role
		}
	}
	logrus.Debugf("AclFileStore.GetRole: clientID=%s, highestRole=%s", clientID, highestRole)
	return highestRole
}

// IsRoleGreaterEqual returns true if a user role has same or greater permissions
// than the minimum role.
func IsRoleGreaterEqual(role string, minRole string) bool {
	if minRole == authorize.GroupRoleNone || role == minRole {
		return true
	}
	if minRole == authorize.GroupRoleViewer && role != authorize.GroupRoleNone {
		return true
	}
	if minRole == authorize.GroupRoleOperator && (role == authorize.GroupRoleManager) {
		return true
	}
	return false
}

// Open the store
// This reads the acl file and subscribes to file changes.
// The ACL file MUST exist, even if it is empty.
func (aclStore *AclFileStore) Open() error {
	logrus.Infof("AclFileStore.Open: clientID='%s'", aclStore.clientID)

	err := aclStore.Reload()
	if err != nil {
		return err
	}
	// watcher handles debounce of too many events
	aclStore.watcher, err = watcher.WatchFile(aclStore.storePath, aclStore.Reload, aclStore.clientID)
	return err
}

// Reload the ACL store from file
func (aclStore *AclFileStore) Reload() error {
	logrus.Infof("AclFileStore.Reload: clientID='%s'. Reloading acls from %s", aclStore.clientID, aclStore.storePath)

	aclStore.mutex.Lock()
	defer aclStore.mutex.Unlock()
	raw, err := os.ReadFile(aclStore.storePath)
	if err != nil {
		logrus.Errorf("AclFileStore.Reload clientID='%s'. File '%s': Error opening the ACL file: %s", aclStore.clientID, aclStore.storePath, err)
		return err
	}
	err = yaml.Unmarshal(raw, &aclStore.Groups)
	if err != nil {
		logrus.Errorf("AclFileStore.Reload clientID='%s'. File '%s': Error parsing the ACL file: %s", aclStore.clientID, aclStore.storePath, err)
		return err
	}

	// update index to lookup the role of a client in a group:
	// eg: [clientID] = {group:role, group:role, ...}
	clientGroups := make(map[string]map[string]string)
	for groupName, group := range aclStore.Groups {
		for client, role := range group {
			cg := clientGroups[client]
			if cg == nil {
				cg = make(map[string]string)
				clientGroups[client] = cg
			}
			cg[groupName] = role
		}
	}
	logrus.Infof("AclFileStore.Reload clientID='%s'. Reloaded %d groups", aclStore.clientID, len(clientGroups))
	aclStore.clientGroups = clientGroups
	return nil
}

// SetRole sets a user ACL and update the store.
// This updates the user's role, saves it to a temp file and move the result to the store file.
// Interruptions will not lead to data corruption as the resulting acl file is only moved after successful write.
// Note that concurrent writes by different processes is not supported and can lead to one of the
// writes being ignored.
//  clientID login name to assign the role
//  groupID  group where the role applies
//  role     one of GroupRoleViewer, GroupRoleOperator, GroupRoleManager, GroupRoleThing or GroupRoleNone to remove the role
func (aclStore *AclFileStore) SetRole(clientID string, groupID string, role string) error {
	// Prevent concurrently running Reload and SetRole
	aclStore.mutex.Lock()
	defer aclStore.mutex.Unlock()

	aclGroup := aclStore.Groups[groupID]
	if aclGroup == nil {
		aclGroup = AclGroup{}
		aclStore.Groups[groupID] = aclGroup
	}
	if role == authorize.GroupRoleNone {
		delete(aclGroup, clientID)
	} else {
		aclGroup[clientID] = role
	}
	folder := path.Dir(aclStore.storePath)
	tempFileName, err := WriteAclsToTempFile(folder, aclStore.Groups)
	if err != nil {
		logrus.Errorf("AclFileStore.SetRole clientID='%s': %s", aclStore.clientID, err)
		return err
	}

	err = os.Rename(tempFileName, aclStore.storePath)
	if err != nil {
		logrus.Errorf("AclFileStore.SetRole. clientID='%s'. Error renaming temp file to store: %s", aclStore.clientID, err)
		return err
	}

	logrus.Infof("AclFileStore.SetRole: clientID='%s' set a role '%s' in group '%s'", clientID, role, groupID)
	return err
}

// Remove removes a client from a group and update the store.
// This removes the user from the given group, or all groups if the 'all' group is used
// Note that concurrent writes by different processes is not supported and can lead to one of the writes being ignored.
//  clientID login name to assign the role
//  groupID  group where the role applies. Use 'all' to remove the user from all groups.
func (aclStore *AclFileStore) Remove(clientID string, groupID string) error {
	// Prevent concurrently running Reload and SetRole
	aclStore.mutex.Lock()
	defer aclStore.mutex.Unlock()

	aclGroup := aclStore.Groups[groupID]
	if aclGroup == nil {
		aclGroup = AclGroup{}
		aclStore.Groups[groupID] = aclGroup
	}
	delete(aclGroup, clientID)
	if groupID == authorize.AclGroupAll {
		// TODO: iterate all groups
		for key := range aclStore.Groups {
			grp := aclStore.Groups[key]
			delete(grp, clientID)
		}
	}

	folder := path.Dir(aclStore.storePath)
	tempFileName, err := WriteAclsToTempFile(folder, aclStore.Groups)
	if err != nil {
		logrus.Errorf("AclFileStore.Remove clientID='%s': %s", aclStore.clientID, err)
		return err
	}

	err = os.Rename(tempFileName, aclStore.storePath)
	if err != nil {
		logrus.Errorf("AclFileStore.Remove. clientID='%s'. Error renaming temp file to store: %s", aclStore.clientID, err)
		return err
	}

	logrus.Infof("AclFileStore.Remove: clientID='%s' removed from group '%s'", clientID, groupID)
	return err
}

// WriteAclsToTempFile write the ACL store to a temp file
func WriteAclsToTempFile(folder string, acls map[string]AclGroup) (tempFileName string, err error) {
	file, err := os.CreateTemp(folder, "hub-aclfilestore")
	// file, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		err := fmt.Errorf("WriteAclsToTempFile: Failed open temp acl file: %s", err)
		return "", err
	}
	tempFileName = file.Name()
	logrus.Infof("WriteAclsToTempFile tempfile: %s", tempFileName)
	defer file.Close()

	data, _ := yaml.Marshal(acls)
	writer := bufio.NewWriter(file)
	_, err = writer.Write(data)
	if err != nil {
		err := fmt.Errorf("WriteAclsToTempFile: Failed writing temp acl file: %s", err)
		return tempFileName, err
	}

	writer.Flush()
	// err := ioutil.WriteFile(path, data, perm)
	return tempFileName, err
}

// NewAclFileStore creates an instance of a file based ACL store
//  filepath is the location of the store. See also DefaultAclFilename for the recommended name.
//  clientID is for logging which authservice is accessing it
func NewAclFileStore(filepath string, clientID string) *AclFileStore {
	store := &AclFileStore{
		clientID:     clientID,
		Groups:       make(map[string]AclGroup),
		clientGroups: make(map[string]map[string]string), // list of groups the client is a member of. Updated on load.

		storePath: filepath,
	}
	return store
}
