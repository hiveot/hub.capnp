package config

// DefaultBackend is the default database type to use
const DefaultBackend = "mongodb"

// DefaultDBName is the name of the storage database
const DefaultDBName = "thinghistory"

// DefaultDBURL holds the URL of the mongodb store
const DefaultDBURL = "mongodb://localhost:27017"

// DefaultDBTimeout to connect to the mongodb server
const DefaultDBTimeout = 3

// HistoryConfig with history store database configuration
type HistoryConfig struct {
	// Name of the database to store
	Backend           string `yaml:"backend"` // only 'mongodb' at the moment
	DatabaseName      string
	DatabaseURL       string
	LoginID           string
	Password          string
	ClientCertificate string // client auth cert
	Timeout           int    // Timeout in seconds to connect to the db
}

// NewHistoryConfig creates a new config with default values
func NewHistoryConfig() HistoryConfig {
	cfg := HistoryConfig{
		Backend:           DefaultBackend,
		ClientCertificate: "",
		DatabaseName:      DefaultDBName,
		DatabaseURL:       DefaultDBURL,
		LoginID:           "",
		Password:          "",
		Timeout:           DefaultDBTimeout,
	}
	return cfg
}
