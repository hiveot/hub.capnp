package idprovpb

import (
	"github.com/wostzone/idprov-go/pkg/idprov"
	"github.com/wostzone/idprov-go/pkg/idprovserver"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

const PluginID = "idprov-pb"

// IDProvPBConfig Protocol binding configuration
type IDProvPBConfig struct {
	IdpAddress      string            `yaml:"idpAddress"`      // listening address, default is automatic
	IdpPort         uint              `yaml:"idpPort"`         // idprov listening port
	IdpCerts        string            `yaml:"idpCerts"`        // folder to store client certificates
	ClientID        string            `yaml:"clientID"`        // unique service instance client ID
	EnableDiscovery bool              `yaml:"enableDiscovery"` // DNS-SD disco
	ValidForDays    uint              `yaml:"validForDays"`    // Nr days certificates are valid for
	Services        map[string]string `yaml:"services"`        // Services that work with provisioned certificates
}

// IDProv provisioning protocol binding service
// Configure and start IDProv server
type IDProvPB struct {
	config    IDProvPBConfig
	hubConfig hubconfig.HubConfig
	server    *idprovserver.IDProvServer
}

// Start the IDProv service
func (pb *IDProvPB) Start() error {

	err := pb.server.Start()
	return err
}

// Stop the IDProv service
func (pb *IDProvPB) Stop() {
	if pb.server != nil {
		pb.server.Stop()
	}
}

// Create a new IDProv protocol binding instance
//  config for IDProv server. Will be updated with defaults
//  hubConfig with certificate info
// Returns IDProv protocol binding instance
func NewIDProvPB(config *IDProvPBConfig, hubConfig *hubconfig.HubConfig) *IDProvPB {
	// use default values if config is incomplete

	// Both mqtt and idprov server must live on the same address to be able to use the same server cert
	if config.IdpAddress == "" {
		// config.IdpAddress = hubconfig.GetOutboundIP("").String()
		config.IdpAddress = hubConfig.MqttAddress
	}
	if config.IdpPort == 0 {
		config.IdpPort = 43776
	}
	if config.IdpCerts == "" {
		config.IdpCerts = "./idpcerts"
	}
	if config.ValidForDays <= 0 {
		config.ValidForDays = 3
	}
	server := idprovserver.NewIDProvServer(
		config.ClientID,
		config.IdpAddress,
		config.IdpPort,
		hubConfig.CertsFolder, config.IdpCerts,
		config.ValidForDays, idprov.IdprovServiceDiscoveryType)

	server.Directory().Services = config.Services

	return &IDProvPB{
		config:    *config,
		hubConfig: *hubConfig,
		server:    server,
	}
}
