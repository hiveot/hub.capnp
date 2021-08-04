package idprovpb

import (
	"path"

	"github.com/wostzone/idprov-go/pkg/idprov"
	"github.com/wostzone/idprov-go/pkg/idprovserver"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

const PluginID = "idprov-pb"
const DefaultCertStore = "./"

// IDProvPBConfig Protocol binding configuration
type IDProvPBConfig struct {
	IdpAddress      string            `yaml:"idpAddress"`      // listening address, default is mqtt server
	IdpPort         uint              `yaml:"idpPort"`         // idprov listening port, default is 43776
	IdpCerts        string            `yaml:"idpCerts"`        // folder to store client certificates
	ClientID        string            `yaml:"clientID"`        // unique service instance client ID
	EnableDiscovery bool              `yaml:"enableDiscovery"` // DNS-SD discovery
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

	// Both mqtt and idprov server live on the same address to use the same server cert
	if config.IdpAddress == "" {
		// config.IdpAddress = hubconfig.GetOutboundIP("").String()
		config.IdpAddress = hubConfig.MqttAddress
	}
	if config.ClientID == "" {
		config.ClientID = PluginID
	}
	if config.IdpPort == 0 {
		config.IdpPort = idprov.DefaultPort
	}
	if config.IdpCerts == "" {
		config.IdpCerts = DefaultCertStore
	}
	if config.ValidForDays <= 0 {
		config.ValidForDays = 3
	}
	discoServiceName := idprov.IdprovServiceName
	if !config.EnableDiscovery {
		discoServiceName = ""
	}
	serverCertPath := path.Join(hubConfig.CertsFolder, certsetup.HubCertFile)
	serverKeyPath := path.Join(hubConfig.CertsFolder, certsetup.HubKeyFile)
	caCertPath := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	caKeyPath := path.Join(hubConfig.CertsFolder, certsetup.CaKeyFile)
	server := idprovserver.NewIDProvServer(
		config.ClientID,
		config.IdpAddress,
		config.IdpPort,
		serverCertPath,
		serverKeyPath,
		caCertPath,
		caKeyPath,
		config.IdpCerts,
		config.ValidForDays,
		discoServiceName)

	server.Directory().Services = config.Services

	return &IDProvPB{
		config:    *config,
		hubConfig: *hubConfig,
		server:    server,
	}
}
