package config

// AuthnConfig contains the authn service configuration
type AuthnConfig struct {
	//
	//// Enable the configuration store for authenticated users. Default is true
	//ConfigStoreEnabled bool `yaml:"configStoreEnabled"`
	//
	//// Set the client config store folder. Default is 'clientconfig' in the config folder
	//ConfigStoreFolder string `yaml:"configStoreFolder"`

	// PasswordFile to read from. Use "" for default defined in 'unpwstore.DefaultPasswordFile'
	PasswordFile string `yaml:"passwordFile"`

	// Access token validity. Default is 1 hour
	AccessTokenValiditySec int `yaml:"accessTokenValiditySec"`

	// Refresh token validity. Default is 1209600 (14 days)
	RefreshTokenValiditySec int `yaml:"refreshTokenValiditySec"`
}
