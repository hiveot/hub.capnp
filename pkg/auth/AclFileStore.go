package auth

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/watcher"
	"gopkg.in/yaml.v3"
)

// The default filename of the ACL file store
const DefaultAclFilename = "groups.acl"

// A group is a map of clients and roles
type AclGroup map[string]string

// AclFileStore stores ACL list in file.
// It includes a file watcher to automatically reload on update.
type AclFileStore struct {
	Groups       map[string]AclGroup `yaml:"groups"`
	storePath    string
	watcher      *fsnotify.Watcher
	mutex        sync.RWMutex
	clientGroups map[string](map[string]string) // list of groups the client is a member of. Updated on load.
}

// Close the store
func (aclStore *AclFileStore) Close() {
	if aclStore.watcher != nil {
		aclStore.watcher.Close()
		aclStore.watcher = nil
	}
}

// GetGroups returns a list of groups a thing or user is a member of
func (aclStore *AclFileStore) GetGroups(clientID string) []string {
	groupsMemberOf := []string{}

	cg := aclStore.clientGroups[clientID]
	for groupName := range cg {
		groupsMemberOf = append(groupsMemberOf, groupName)
	}
	return groupsMemberOf
}

// Get highest role of a user has in a list of group
// Intended to get client permissions in case of overlapping groups
func (aclStore *AclFileStore) GetRole(clientID string, groupIDs []string) string {
	highestRole := GroupRoleNone

	cg := aclStore.clientGroups[clientID]

	for _, groupID := range groupIDs {
		role := cg[groupID]
		if IsRoleGreaterEqual(role, highestRole) {
			highestRole = role
		}
	}
	return highestRole
}

// IsRoleGreaterEqual returns true if a user role has same or greater permissions
// than the minimum role.
func IsRoleGreaterEqual(role string, minRole string) bool {
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
// This reads the acl file and subscribes to file changes.
// The ACL file MUST exist, even if it is empty.
func (aclStore *AclFileStore) Open() error {
	err := aclStore.Reload()
	if err != nil {
		return err
	}
	aclStore.watcher, err = watcher.WatchFile(aclStore.storePath, aclStore.Reload)
	return err
}

// reload the ACL store from file
func (aclStore *AclFileStore) Reload() error {
	aclStore.mutex.RLock()
	defer aclStore.mutex.RUnlock()
	logrus.Infof("AclFileStore.Reload: Reloading acls from %s", aclStore.storePath)

	raw, err := os.ReadFile(aclStore.storePath)
	if err != nil {
		logrus.Errorf("AclFileStore.Reload '%s': Error opening the ACL file: %s", aclStore.storePath, err)
		return err
	}
	err = yaml.Unmarshal(raw, aclStore)
	if err != nil {
		logrus.Errorf("AclFileStore.Reload '%s': Error parsing the ACL file: %s", aclStore.storePath, err)
		return err
	}

	// update index to lookup the role of a client in a group:
	// eg: [clientID] = {group:role, group:role, ...}
	clientGroups := make(map[string](map[string]string))
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
	logrus.Infof("AclFileStore.Reloaded %d groups", len(clientGroups))
	aclStore.clientGroups = clientGroups
	return nil
}

// Set a user ACL and update the store.
// This updates the user's role, saves it to a temp file and move the result to the store file.
// Interruptions will not lead to data corruption as the resulting acl file is only moved after successful write.
// Note that concurrent writes by different processes is not supported and can lead to one of the
// writes being ignored.
//  clientID login name to assign the role
//  groupID  group where the role applies
//  role     one of GroupRoleViewer, GroupRoleEditor, GroupRoleManager, GroupRoleThing or GroupRoleNone to remove the role
func (aclStore *AclFileStore) SetRole(clientID string, groupID string, role string) error {
	// Prevent concurrently running Reload and SetRole
	aclStore.mutex.Lock()
	defer aclStore.mutex.Unlock()

	aclGroup := aclStore.Groups[groupID]
	if aclGroup == nil {
		aclGroup = AclGroup{}
		aclStore.Groups[groupID] = aclGroup
	}
	if role == GroupRoleNone {
		delete(aclGroup, clientID)
	} else {
		aclGroup[clientID] = role
	}
	folder := path.Dir(aclStore.storePath)
	tempFileName, err := aclStore.WriteToTemp(folder)
	if err != nil {
		logrus.Infof("AclFileStore.SetRole, error writing to temp file: %s", err)
		return err
	}

	// tmpPath := aclStore.storePath + ".tmp"
	// err := aclStore.Write(tmpPath, 0600)
	// if err != nil {
	// 	return err
	// }
	err = os.Rename(tempFileName, aclStore.storePath)
	if err != nil {
		logrus.Infof("AclFileStore.SetRole, error renaming temp file to store: %s", err)
		return err
	}

	logrus.Infof("Set: Client '%s' set a role '%s' in group '%s'", clientID, role, groupID)
	return err
}

// Write the ACL store to a temp file
func (aclStore *AclFileStore) WriteToTemp(folder string) (tempFileName string, err error) {
	file, err := os.CreateTemp(folder, "hub-aclfilestore")
	// file, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		err := fmt.Errorf("AclFileStore.Write: Failed open temp acl file: %s", err)
		// logrus.Error(err)
		return "", err
	}
	tempFileName = file.Name()
	defer file.Close()

	data, _ := yaml.Marshal(aclStore)
	writer := bufio.NewWriter(file)
	_, err = writer.Write(data)
	if err != nil {
		err := fmt.Errorf("AclFileStore.Write: Failed writing temp acl file: %s", err)
		// logrus.Error(err)
		return tempFileName, err
	}

	writer.Flush()
	// err := ioutil.WriteFile(path, data, perm)
	return tempFileName, err
}

// New instance of a file based ACL store
//  filepath is the location of the store. See also DefaultAclFilename for the recommended name.
func NewAclFileStore(filepath string) *AclFileStore {
	store := &AclFileStore{
		Groups:    map[string]AclGroup{},
		storePath: filepath,
	}
	return store
}
