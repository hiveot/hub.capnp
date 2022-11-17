package config

import "github.com/hiveot/hub/pkg/bucketstore"

// DefaultBackend is the default database type to use
const DefaultBackend = bucketstore.BackendPebble

// HistoryConfig with history store database configuration
type HistoryConfig struct {
	// Name of the database to store
	Backend   string `yaml:"backend"`   // kvbtree, pebble, bbolt. See IBucketStore for options.
	Directory string `yaml:"directory"` // backend directory
}

// NewHistoryConfig creates a new config with default values
func NewHistoryConfig(storeDirectory string) HistoryConfig {
	cfg := HistoryConfig{
		Backend:   DefaultBackend,
		Directory: storeDirectory,
	}
	return cfg
}
