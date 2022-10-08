package config

// DefaultDBName is the name of the storage database
const DefaultDBName = "historystore"

// DefaultDBType is the default database type to use
const DefaultDBType = "mongodb"

// DefaultDBURL holds the URL of the mongodb store
const DefaultDBURL = "mongodb://localhost:27017"

// HistoryStoreConfig with history store database configuration
type HistoryStoreConfig struct {
	// Name of the database to store
	DatabaseType    string // only 'mongodb' at the moment
	DatabaseName    string
	DatabaseURL     string
	LoginID         string
	Password        string
	CertificateFile string // client auth cert
}

// NewHistoryStoreConfig creates a new config with default values
func NewHistoryStoreConfig() HistoryStoreConfig {
	cfg := HistoryStoreConfig{
		DatabaseType:    DefaultDBType,
		DatabaseName:    DefaultDBName,
		DatabaseURL:     DefaultDBURL,
		LoginID:         "",
		Password:        "",
		CertificateFile: "",
	}
	return cfg
}
