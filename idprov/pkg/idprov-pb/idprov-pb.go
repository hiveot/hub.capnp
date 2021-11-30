// Package idprovpb with the protocol binding between wost hub configuration and the IDProv server
package idprovpb

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/hub/idprov/pkg/idprovserver"
)

const PluginID = "idprov-pb"
const DefaultCertStore = "./"
const mqttSSLServiceName = "mqtts"
const mqttWSServiceName = "mqttws"

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

// IDProvPB provisioning protocol binding service
// Configure and start IDProv server
type IDProvPB struct {
	config IDProvPBConfig
	// hubConfig config.HubConfig
	server *idprovserver.IDProvServer
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

// NewIDProvPB creates a new IDProv protocol binding instance
//  config for IDProv server. Will be updated with defaults
//  address of the server to advertise
//  mqttCertPort of the mqtt broker for mutual certificate auth clients
//  mqttWSPort of the mqtt broker for websocket clients
//  serverCert TLS certificate of the IDProv server, signed by the CA
//  caCert CA certificate used to sign client certificates
//  caKey CA's key used to sign client certificates
// Returns IDProv protocol binding instance
func NewIDProvPB(config *IDProvPBConfig,
	address string,
	mqttCertPort uint,
	mqttWSPort uint,
	serverCert *tls.Certificate,
	caCert *x509.Certificate,
	caKey *ecdsa.PrivateKey) *IDProvPB {
	// use sane default values

	// Both mqtt and idprov server live on the same address to use the same server cert
	if config.IdpAddress == "" {
		config.IdpAddress = address
	}
	if config.ClientID == "" {
		config.ClientID = PluginID
	}
	if config.IdpPort == 0 {
		config.IdpPort = idprovclient.DefaultPort
	}
	if config.IdpCerts == "" {
		config.IdpCerts = DefaultCertStore
	}
	if config.ValidForDays <= 0 {
		config.ValidForDays = 3
	}
	discoServiceName := idprovclient.IdprovServiceName
	if !config.EnableDiscovery {
		discoServiceName = ""
	}
	// include the location of the mqtt server for certificate and unpw/websocket connections
	if config.Services == nil {
		config.Services = make(map[string]string)
	}
	config.Services[mqttSSLServiceName] = fmt.Sprintf("tls://%s:%d", address, mqttCertPort)
	config.Services[mqttWSServiceName] = fmt.Sprintf("wss://%s:%d", address, mqttWSPort)

	// serverCertPath := path.Join(hubConfig.CertsFolder, certsetup.HubCertFile)
	// serverKeyPath := path.Join(hubConfig.CertsFolder, certsetup.HubKeyFile)
	// caCertPath := path.Join(hubConfig.CertsFolder, certsetup.CaCertFile)
	// caKeyPath := path.Join(hubConfig.CertsFolder, certsetup.CaKeyFile)
	server := idprovserver.NewIDProvServer(
		config.ClientID,
		config.IdpAddress,
		config.IdpPort,
		serverCert,
		caCert,
		caKey,
		config.IdpCerts,
		config.ValidForDays,
		discoServiceName)

	// todo: template substitution of things like address
	server.Directory().Services = config.Services

	return &IDProvPB{
		config: *config,
		// hubConfig: *hubConfig,
		server: server,
	}
}
