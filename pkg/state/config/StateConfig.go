package config

import "github.com/hiveot/hub/pkg/bucketstore"

// StateConfig holds the configuration of the state service
type StateConfig struct {
	// Embedded store backend: options are 'kvstore' | 'bbolt' | 'pebble' (default)
	Backend string `yaml:"backend"`

	// Directory where DB files and folders are stored
	StoreDirectory string `yaml:"storeDirectory"`

	// Constraints on storing state for services
	Services struct {
		MaxKeys      int `yaml:"maxKeys"`
		MaxValueSize int `yaml:"maxValueSize"`
	} `yaml:"services"`

	// Constraints on storing state for end-users
	Users struct {
		MaxKeys      int `yaml:"maxKeys"`
		MaxValueSize int `yaml:"maxValueSize"`
	} `yaml:"users"`
}

// NewStateConfig returns a new configuration with defaults
func NewStateConfig(storeDirectory string) StateConfig {
	sc := StateConfig{}
	sc.Backend = bucketstore.BackendKVBTree
	sc.StoreDirectory = storeDirectory
	sc.Services.MaxKeys = 100
	sc.Services.MaxValueSize = 100000
	sc.Users.MaxKeys = 100
	sc.Users.MaxValueSize = 100000
	return sc
}
