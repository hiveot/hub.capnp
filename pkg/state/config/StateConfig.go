package config

// Embedded stores the state service can use
const (
	// StateBackendKVStore is an in-memory store that is super fast but limited to memory.
	// Data stored in a single file.
	// Easy to backup. Good for many small clients.
	StateBackendKVStore = "kvstore"

	// StateBackendBBolt is a boltDB compatible store. Fast on read but slow on write.
	// DB can grow beyond available memory but writes slow down as the data increase beyond GB.
	// Stores all data in a single file.
	// Easy to backup. Good for compatibility with existing bbolt db.
	StateBackendBBolt = "bbolt"

	// StateBackendPebble is CockroachDB's pebble backend. Fast on read and write.
	// DB can grow beyond many GB without significant slowdown.
	// Uses a folder with various files.
	// Need tool to backup. (tbd) Best for large datasets.
	StateBackendPebble = "pebble"
)

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
func NewStateConfig(StoreDirectory string) StateConfig {
	sc := StateConfig{}
	sc.Backend = StateBackendPebble
	sc.StoreDirectory = StoreDirectory
	sc.Services.MaxKeys = 100
	sc.Services.MaxValueSize = 100000
	sc.Users.MaxKeys = 100
	sc.Users.MaxValueSize = 100000
	return sc
}
