package config

import (
	"path"

	"github.com/hiveot/hub/pkg/state"
)

// StateConfig holds the configuration of the state service
type StateConfig struct {
	// Store backend: options are 'kvstore'
	Backend string `yaml:"backend"`
	// Database name if applicable
	DatabaseName string `yaml:"databaseName"`
	// Database URL or store file. Default is state.json
	DatabaseURL string `yaml:"databaseURL"`

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
func NewStateConfig(storesFolder string) StateConfig {
	sc := StateConfig{}
	sc.Backend = "kvstore"
	sc.DatabaseName = "hubstate"
	sc.DatabaseURL = path.Join(storesFolder, state.ServiceName+".json") // in the stores folder
	sc.Services.MaxKeys = 100
	sc.Services.MaxValueSize = 100000
	sc.Users.MaxKeys = 100
	sc.Users.MaxValueSize = 100000
	return sc
}
