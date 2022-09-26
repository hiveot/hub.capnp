package idprovserver

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/wostzone/wost-go/pkg/hubnet"
	"net/http"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/wost-go/pkg/tlsserver"
)

const PluginID = "idprov"

// routes for API access
const (
	RouteGetDirectory         = idprovclient.IDProvDirectoryPath
	RouteGetDeviceStatus      = "/idprov/status/{deviceID}"
	RoutePostOOB              = "/idprov/oobsecret"
	RoutePostProvisionRequest = "/idprov/provreq"
)

// myDirectory is used to initialize the server with
var myDirectory = idprovclient.GetDirectoryMessage{
	Endpoints: idprovclient.DirectoryEndpoints{
		GetDirectory:            RouteGetDirectory,
		PostProvisioningRequest: RoutePostProvisionRequest,
		GetDeviceStatus:         RouteGetDeviceStatus,
		PostOobSecret:           RoutePostOOB,
	},
	Services:  map[string]string{},
	CaCertPEM: nil,
	Version:   "1",
}

// IDProvConfig server configuration
type IDProvConfig struct {

	// Nr days issued certificates are valid for.
	// Defaults to 7 days.
	CertValidityDays uint `yaml:"certValidityDays"`

	// CertStoreFolder is the location where the server stores issued client certificates
	// Default is "" which means issued certificates are not stored
	CertStoreFolder string `yaml:"certStoreFolder"`

	// Unique Client ID of this instance.
	// Used in service identification on the message bus and in discovery.
	// Default is the plugin ID 'idprov'
	ClientID string `yaml:"clientID"`

	// DisableDiscovery disables publishing of DNS-SD discovery records
	// Default is false (enabled)
	DisableDiscovery bool `yaml:"disableDiscovery"`

	// IdpAddress contains the listening address.
	// Default is the outbound interface address on the local subnet.
	IdpAddress string `yaml:"idpAddress"`

	// idprov listening port
	// Default is idprovclient.DefaultPort
	IdpPort uint `yaml:"idpPort"`

	// ServiceName to publish in discovery.
	// Default is 'idprovclient.IdprovServiceName'
	ServiceName string `yaml:"serviceName"`
}

// IDProvServer runs the server for the IoT provisioning protocol.
//
// This verifies the device secret with the out-of-bound provided secret and issues an IoT device
// client certificate, signed by the CA.
//
// If enabled, a discovery record is published using DNS-SD to allow potential clients to find the
// address and ports of the provisioning server, and optionally additional services.
type IDProvServer struct {
	// service configuration
	config *IDProvConfig

	//address    string // listening address
	directory  idprovclient.GetDirectoryMessage
	serverCert *tls.Certificate
	caCert     *x509.Certificate
	caKey      *ecdsa.PrivateKey
	// caKeyPEM   []byte // loaded CA key PEM
	//instanceID string // unique service ID for discovery and connections

	// runtime status
	running   bool
	tlsServer *tlsserver.TLSServer
	// router      *mux.Router
	discoServer *zeroconf.Server

	oobSecrets map[string]string // [deviceID]secret simple in-memory store for OOB secrets
}

// Return the address that the server listens on
// func (srv *IDProvServer) Address() string {
// 	return srv.address
// }

// Directory returns the directory that the server publishes
// The directory can be modified
func (srv *IDProvServer) Directory() *idprovclient.GetDirectoryMessage {
	return &srv.directory
}

// ServeDirectory returns the endpoint directory of the server
// This method does not require authentication
func (srv *IDProvServer) ServeDirectory(resp http.ResponseWriter, req *http.Request) {
	logrus.Infof("IdProvServer.ServeDirectory")
	msg, _ := json.Marshal(srv.directory)
	_, _ = resp.Write(msg)

}

// Start the IdProv server.
// On systems with multiple addresses, the address with the default outbound interface is used.
// If DNS-SD discovery cannot be started at this time then it can be started later with ServeDiscovery()
func (srv *IDProvServer) Start() error {
	var err error

	if srv.caKey == nil || srv.caCert == nil {
		err := fmt.Errorf("IDProvServer.Start: Missing parameters")
		logrus.Error(err)
		return err
	}
	if !srv.running {
		// srv.listenAddress = listenAddress
		srv.running = true

		logrus.Warningf("Starting IdProv server on %s:%d", srv.config.IdpAddress, srv.config.IdpPort)
		// srv.directory.CaCertPEM, err = ioutil.ReadFile(srv.caCertPath)
		// if err != nil {
		// 	logrus.Errorf("IDProvServer.Start: Loading CA Certificate failed: %s", err)
		// 	return err
		// }
		// srv.caKeyPEM, err = ioutil.ReadFile(srv.caKeyPath)
		// if err != nil {
		// 	logrus.Errorf("IDProvServer.Start: Loading CA Key failed: %s", err)
		// 	return err
		// }

		// do not use the authenticator as the trust is still to be established
		srv.tlsServer = tlsserver.NewTLSServer(srv.config.IdpAddress, srv.config.IdpPort,
			srv.serverCert, srv.caCert)
		err := srv.tlsServer.Start()
		if err != nil {
			logrus.Errorf("IDProvServer.Start: Failed starting IDProv server: %s", err)
			return err
		}
		// setup the handlers for the paths
		srv.tlsServer.AddHandlerNoAuth(RouteGetDirectory, srv.ServeDirectory)
		srv.tlsServer.AddHandler(RouteGetDeviceStatus, srv.ServeStatus)
		srv.tlsServer.AddHandler(RoutePostOOB, srv.ServePostOOB)
		srv.tlsServer.AddHandlerNoAuth(RoutePostProvisionRequest, srv.ServeProvisionRequest)

		if !srv.config.DisableDiscovery {
			srv.discoServer, err = srv.ServeIdProvDiscovery(srv.config.ServiceName)
			if err != nil {
				logrus.Errorf("IdProvServer.Start: failed starting discovery: %s (continuing without)", err)
				err = nil
			}
		}
		// Make sure the server is listening before continuing
		// Not pretty but it handles it
		time.Sleep(time.Second)
	}
	return err
}

//Stop the IdProv server
func (srv *IDProvServer) Stop() {
	if srv.running {
		srv.tlsServer.Stop()
		srv.running = false
		if srv.discoServer != nil {
			srv.discoServer.Shutdown()
			srv.discoServer = nil
		}
	}
}

// NewIDProvServer creates a new instance of the IoT Device Provisioning Server
//  instanceID is the unique ID for this service used in discovery and communication
//  config with server configuration settings.
//  serverCert, server own TLS certificate
//  caCert CA x509 Certificate
//  caKey CA private/public Key PEM file needed to issue signed certificates
//  serviceName with the discovery services name. Use "" to disable discover, or idprovclient.IdprovServiceName for default
func NewIDProvServer(
	config *IDProvConfig,
	serverCert *tls.Certificate,
	caCert *x509.Certificate,
	caKey *ecdsa.PrivateKey,
) *IDProvServer {

	if config.ClientID == "" {
		config.ClientID = PluginID
	}
	if config.IdpPort == 0 {
		config.IdpPort = idprovclient.DefaultPort
	}
	if config.IdpAddress == "" {
		config.IdpAddress = hubnet.GetOutboundIP("").String()
	}
	if config.CertValidityDays <= 0 {
		config.CertValidityDays = 30
	}
	if config.ServiceName == "" {
		config.ServiceName = idprovclient.IdprovServiceName
	}

	srv := IDProvServer{
		config:     config,
		serverCert: serverCert,
		caCert:     caCert,
		caKey:      caKey,
		directory:  myDirectory,
		oobSecrets: make(map[string]string),
	}
	return &srv
}
