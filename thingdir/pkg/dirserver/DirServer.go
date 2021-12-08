// Package dirserver for serving access to the directory store
package dirserver

import (
	"crypto/tls"
	"crypto/x509"
	"path"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/auth/pkg/authenticate"
	"github.com/wostzone/hub/auth/pkg/authorize"
	"github.com/wostzone/hub/lib/serve/pkg/tlsserver"
	"github.com/wostzone/hub/thingdir/pkg/dirclient"
	"github.com/wostzone/hub/thingdir/pkg/dirstore/dirfilestore"
)

const DirectoryPluginID = "directory"
const DefaultDirectoryStoreFile = "directory.json"

// const RouteUpdateTD = "/things/{thingID}"
// const RouteGetTD = "/things/{thingID}"
// const RouteDeleteTD = "/things/{thingID}"
// const RoutePatchTD = "/things/{thingID}"
// const RouteListTD = "/things"
// const RouteQueryTD = "/things"

// DirectoryServer for web of things
type DirectoryServer struct {
	// config
	address       string            // listening address
	caCert        *x509.Certificate // path to CA certificate PEM file
	instanceID    string            // ID of this service
	port          uint              // listening port
	serverCert    *tls.Certificate  // path to server certificate PEM file
	authenticator authenticate.VerifyUsernamePassword
	authorizer    authorize.VerifyAuthorization

	// the service name. Use dirclient.DirectoryServiceName for default or "" to disable DNS discovery
	discoveryName string

	// runtime status
	running     bool
	tlsServer   *tlsserver.TLSServer
	discoServer *zeroconf.Server
	store       *dirfilestore.DirFileStore
}

// Address returns the address that the server listens on
// This is automatically determined from the default network interface
func (srv *DirectoryServer) Address() string {
	return srv.address
}

// Start the server.
func (srv *DirectoryServer) Start() error {
	var err error

	if !srv.running {
		srv.running = true

		logrus.Warningf("Starting directory server on %s:%d", srv.address, srv.port)

		// load the saved directory content from file
		err = srv.store.Open()
		if err != nil {
			return err
		}

		// srv.address = hubconfig.GetOutboundIP("").String()
		srv.tlsServer = tlsserver.NewTLSServer(srv.address, srv.port,
			srv.serverCert, srv.caCert)
		srv.tlsServer.EnableJwtAuth(nil)

		// Allow login on this server if an authenticator is provided
		if srv.authenticator != nil {
			srv.tlsServer.EnableJwtIssuer(nil, srv.authenticator)
			//srv.tlsServer.EnableBasicAuth(srv.authenticator)
		}

		err = srv.tlsServer.Start()
		if err != nil {
			return err
		}

		// setup the handlers for the paths. The GET/PUT/... operations are resolved by the handler
		srv.tlsServer.AddHandler(dirclient.RouteThings, srv.ServeThings)
		srv.tlsServer.AddHandler(dirclient.RouteThingID, srv.ServeThingByID)

		// DNS-SD service discovery is optional
		if srv.discoveryName != "" {
			srv.discoServer, _ = ServeDirDiscovery(srv.instanceID, srv.discoveryName, srv.address, srv.port)
		}
		// Make sure the server is listening before continuing
		// Not pretty but it handles it
		time.Sleep(time.Second)
	}
	return nil
}

// Stop the directory server
func (srv *DirectoryServer) Stop() {
	if srv.running {
		srv.running = false
		logrus.Warningf("Stopping directory server on %s:%d", srv.address, srv.port)
		if srv.discoServer != nil {
			srv.discoServer.Shutdown()
			srv.discoServer = nil
		}
		if srv.tlsServer != nil {
			srv.tlsServer.Stop()
			srv.tlsServer = nil
		}
		srv.store.Close()

	}
}

// NewDirectoryServer creates a new instance of the IoT Device Provisioning Server.
//  - instanceID is the unique ID for this service used in discovery and communication
//  - storeFolder is the location of the directory storage file. This must be writable.
//  - address the server listening address. Typically the same address as the mqtt bus
//  - port server listening port
//  - caCertFolder location of CA Cert and server certificates and keys
//  - discoveryName for use in dns-sd. Use "" to disable discover, or the dirclient.DirectoryServiceName for default
//  - authenticator authenticates user login and issues JWT tokens. nil to use an external auth service
//  - authorizer verifies read or write access to a thing by a user. certOU is set when auth via certificate
func NewDirectoryServer(
	instanceID string,
	storeFolder string,
	address string,
	port uint,
	discoveryName string,
	serverCert *tls.Certificate,
	caCert *x509.Certificate,
	authenticator authenticate.VerifyUsernamePassword,
	authorizer authorize.VerifyAuthorization,
) *DirectoryServer {

	if instanceID == "" || port == 0 {
		logrus.Panic("NewDirectoryServer: Invalid arguments for instanceID or port")
		panic("Exit due to invalid args")
	}
	storePath := path.Join(storeFolder, DefaultDirectoryStoreFile)
	srv := DirectoryServer{
		address:       address,
		serverCert:    serverCert,
		caCert:        caCert,
		discoveryName: discoveryName,
		instanceID:    instanceID,
		port:          port,
		store:         dirfilestore.NewDirFileStore(storePath),
		authenticator: authenticator,
		authorizer:    authorizer,
	}
	return &srv
}