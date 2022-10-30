package clientconfigstore

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// ClientConfigStore storage for storing user configuration snippets
// This provides authenticated users a place to store their app configuration.
type ClientConfigStore struct {
	// folder containing the configuration files per user
	storeFolder string
}

// Close the store
func (cfgStore *ClientConfigStore) Close() {
	logrus.Infof("ClientConfigStore.Close")
}

// Get user application config from the store
// Returns a string with configuration text or an empty string if the store doesn't exist
func (cfgStore *ClientConfigStore) Get(userID, appID string) string {
	logrus.Infof("ClientConfigStore.Get userID='%s', appID='%s'", userID, appID)
	cfgFileName := fmt.Sprintf("%s.%s.cfg", userID, appID)
	cfgFilePath := path.Join(cfgStore.storeFolder, cfgFileName)
	configText, err := ioutil.ReadFile(cfgFilePath)
	if err != nil {
		logrus.Infof("ClientConfigStore.Get: %s", err)
		return ""
	}
	return string(configText)
}

// Put user application configuration into the store.
// This writes the configText to a file <userID>.<appID>.cfg in the store folder
//
//  userID is the ID of the user whose store to update
//  appID is the configuration application ID
//  configText is the configuration in text format
func (cfgStore *ClientConfigStore) Put(userID, appID string, configText string) error {
	logrus.Infof("ClientConfigStore.Put")
	cfgFileName := fmt.Sprintf("%s.%s.cfg", userID, appID)
	cfgFilePath := path.Join(cfgStore.storeFolder, cfgFileName)
	err := ioutil.WriteFile(cfgFilePath, []byte(configText), 0600)
	return err
}

// Open the store
// Create the folder if it doesn't exist
func (cfgStore *ClientConfigStore) Open() error {
	var err error
	// Right now open doesn't do much, except creating the folder if needed.
	// Future improvements might use a sqlite or other database type solution
	logrus.Infof("ClientConfigStore.Open. location='%s'", cfgStore.storeFolder)
	if _, err := os.Stat(cfgStore.storeFolder); err != nil {
		err = os.MkdirAll(cfgStore.storeFolder, 0755)
	}
	if err != nil {
		return err
	}
	return err
}

func NewClientConfigStore(storeFolder string) *ClientConfigStore {
	cfgFolder := &ClientConfigStore{storeFolder: storeFolder}
	return cfgFolder
}
