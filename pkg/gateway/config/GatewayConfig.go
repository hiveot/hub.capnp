package config

import (
	"fmt"

	"github.com/hiveot/hub/lib/listener"
)

const DefaultGatewayPort = ":8883"
const DefaultGatewayWSPort = ":8884/ws"

type GatewayConfig struct {
	// server listening address:port
	Address string `yaml:"address"`

	// websocket  listening address:port/path. Default is {Address}/ws
	WSAddress string `yaml:"wsAddress"`

	// noTLS disables TLS. Default is enabled. Intended for testing.
	NoTLS bool `yaml:"noTLS"`

	// useWS disables the websocket listener
	NoWS bool `yaml:"noWS"`
}

// NewGatewayConfig creates a new gateway configuration with defaults
func NewGatewayConfig() *GatewayConfig {
	oip := listener.GetOutboundIP("")
	gwConfig := GatewayConfig{
		Address:   fmt.Sprintf("%s%s", oip.String(), DefaultGatewayPort),
		WSAddress: fmt.Sprintf("%s%s", oip.String(), DefaultGatewayWSPort),
	}
	return &gwConfig
}
