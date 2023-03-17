package config

import (
	"github.com/hiveot/hub/lib/listener"
)

const DefaultGatewayTcpPort = 9883 // TLS over TCP
const DefaultGatewayWssPort = 9884 // Websocket over TLS
const DefaultGatewayWssPath = "/ws"

type GatewayConfig struct {
	// server listening address or "" for automatic outbound IP
	Address string `yaml:"address"`

	// noDiscovery disables the DNS-SD discovery
	// discovery is useful for remote clients failover services
	NoDiscovery bool `yaml:"noDiscovery"`

	// noTLS disables TLS. Default is enabled. Intended for testing.
	NoTLS bool `yaml:"noTLS,omitempty"`

	// noWS disables websockets. Default is enabled. Intended for testing.
	NoWS bool `yaml:"noWS,omitempty"`

	// TCP listening port, default is 9883
	TcpPort int `yaml:"tcpPort"`

	// websocket listening port, default is 9884
	WssPort int `yaml:"wssPort,omitempty"`

	// websocket listening path, default is "/ws"
	WssPath string `yaml:"wssPath,omitempty"`
}

// NewGatewayConfig creates a new gateway configuration with defaults
func NewGatewayConfig() *GatewayConfig {
	oip := listener.GetOutboundIP("")
	gwConfig := GatewayConfig{
		Address: oip.String(),
		TcpPort: DefaultGatewayTcpPort,
		WssPort: DefaultGatewayWssPort,
		WssPath: DefaultGatewayWssPath,
	}
	return &gwConfig
}
