package config

const DefaultGatewayAddress = "127.0.0.1:8884"

type GatewayConfig struct {
	// server listening address:port
	Address string `json:"address"`

	// noTLS disables TLS. Default is enabled. Intended for testing.
	NoTLS bool `json:"noTLS"`

	// location of services sockets. Default is {home}/run.
	SocketFolder string `json:"socketFolder"`

	// AutoStart automatically launches requested services if they are not running. Default false.
	AutoStart bool `json:"autoStart"`
}

// NewGatewayConfig creates a new gateway configuration with defaults
func NewGatewayConfig(socketFolder string, certsFolder string) *GatewayConfig {
	gwConfig := GatewayConfig{
		Address:      DefaultGatewayAddress,
		SocketFolder: socketFolder,
	}
	return &gwConfig
}
