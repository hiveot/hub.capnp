package config

import (
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/history"
)

// DefaultBackend is the default database type to use
const DefaultBackend = bucketstore.BackendPebble

// HistoryConfig with history store database configuration
type HistoryConfig struct {
	// Name of the backend to store
	// kvbtree, pebble (default), bbolt. See IBucketStore for options.
	Backend string `yaml:"backend"`

	// Location where to store the history
	Directory string `yaml:"directory"`

	// instance ID of the service, eg: "history". urn: prefix will be added when used as thingID
	ServiceID string `yaml:"serviceID"`
}

// NewHistoryConfig creates a new config with default values
func NewHistoryConfig(storeDirectory string) HistoryConfig {
	cfg := HistoryConfig{
		Backend:   DefaultBackend,
		Directory: storeDirectory,
		ServiceID: history.ServiceName,
	}
	return cfg
}
