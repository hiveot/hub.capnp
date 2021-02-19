package config

// LoadConfig parses the commandline and loads the specified configuration
// Commandline arguments are applied to the configuration. Predefined:
//  --config=plugin configfile
//  --host=host:port of messaging server
//  --usetls=certificate folder
//
// Returns nil if successful
func LoadConfig(config interface{}) error {
	//
	return nil
}

// ParseCommandline and return the map with key-value
func ParseCommandline() map[string]string {

	args := make(map[string]string)
	return args
}
