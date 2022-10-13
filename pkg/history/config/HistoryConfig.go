package config

// DefaultBackend is the default database type to use
const DefaultBackend = "mongodb"

// DefaultDBName is the name of the storage database
const DefaultDBName = "thinghistory"

// DefaultDBURL holds the URL of the mongodb store
const DefaultDBURL = "mongodb://localhost:27017"

// HistoryConfig with history store database configuration
type HistoryConfig struct {
	// Name of the database to store
	Backend           string `yaml:"backend"` // only 'mongodb' at the moment
	DatabaseName      string
	DatabaseURL       string
	LoginID           string
	Password          string
	ClientCertificate string // client auth cert
}

// NewHistoryConfig creates a new config with default values
func NewHistoryConfig() HistoryConfig {
	cfg := HistoryConfig{
		Backend:           DefaultBackend,
		DatabaseName:      DefaultDBName,
		DatabaseURL:       DefaultDBURL,
		LoginID:           "",
		Password:          "",
		ClientCertificate: "",
	}
	return cfg
}
