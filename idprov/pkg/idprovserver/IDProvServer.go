package idprovserver

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/wost-go/pkg/tlsserver"
)

const RouteGetDirectory = idprovclient.IDProvDirectoryPath
const RouteGetDeviceStatus = "/idprov/status/{deviceID}"
const RoutePostOOB = "/idprov/oobsecret"
const RoutePostProvisionRequest = "/idprov/provreq"

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

type IDProvServer struct {
	// config
	address    string // listening address
	directory  idprovclient.GetDirectoryMessage
	serverCert *tls.Certificate
	caCert     *x509.Certificate
	caKey      *ecdsa.PrivateKey
	// caKeyPEM   []byte // loaded CA key PEM
	//
	certStore              string // folder where generated client certificates are stored
	deviceCertValidityDays uint   // nr of days a new certificate is valid for
	port                   uint   // listening port
	// the discovery service name. Use "" to disable discovery or idprovclient.IdprovServiceName for default
	serviceName string
	instanceID  string // unique service ID for discovery and connections

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

	if srv.instanceID == "" || srv.port == 0 || srv.caKey == nil || srv.caCert == nil {
		err := fmt.Errorf("IDProvServer.Start: Missing parameters")
		logrus.Error(err)
		return err
	}
	if !srv.running {
		// srv.listenAddress = listenAddress
		srv.running = true

		logrus.Warningf("Starting IdProv server on %s:%d", srv.address, srv.port)
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
		srv.tlsServer = tlsserver.NewTLSServer(srv.address, srv.port,
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

		if srv.serviceName != "" {
			srv.discoServer, err = srv.ServeIdProvDiscovery(srv.serviceName)
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
//  address the server listening address. Must use the same address as the services
//  port server listening port
//  serverCert, server own TLS certificate
//  caCert CA x509 Certificate
//  caKey CA private/public Key PEM file needed to issue signed certificates
//  certStore location of client certificate storage
//  certValidityDays nr of days device certificate are valid for
//  clientFolder location of generated client certificates
//  serviceName with the discovery services name. Use "" to disable discover, or idprovclient.IdprovServiceName for default
func NewIDProvServer(
	instanceID string,
	address string,
	port uint,
	serverCert *tls.Certificate,
	caCert *x509.Certificate,
	caKey *ecdsa.PrivateKey,
	certStore string,
	certValidityDays uint,
	serviceName string) *IDProvServer {

	srv := IDProvServer{
		address:                address,
		serverCert:             serverCert,
		caCert:                 caCert,
		caKey:                  caKey,
		certStore:              certStore,
		deviceCertValidityDays: certValidityDays,
		directory:              myDirectory,
		serviceName:            serviceName,
		instanceID:             instanceID,
		oobSecrets:             make(map[string]string),
		port:                   port,
	}
	return &srv
}
