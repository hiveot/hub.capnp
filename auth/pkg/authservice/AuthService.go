// Package authservice to serve a REST api for authentication and token refresh
package authservice

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/auth/pkg/authenticate"
	"github.com/wostzone/hub/auth/pkg/configstore"
	"github.com/wostzone/hub/auth/pkg/unpwstore"
	"github.com/wostzone/hub/lib/client/pkg/tlsclient"
	"github.com/wostzone/hub/lib/serve/pkg/tlsserver"
)

// PluginID of the service
const PluginID = "authservice"

// DefaultAuthServicePort to connect to the auth service
const DefaultAuthServicePort = 8881

// internal constant for appID route parameter
const appIDParam = "appid"

// AuthServiceConfig contains the service configuration
type AuthServiceConfig struct {
	// Override of the server address from hub.yaml
	Address string `yaml:"address"`

	// Service listening port or 0 to use the default port 8881
	Port uint `yaml:"port"`

	// ClientID to identify this service as. Default is the pluginID
	ClientID string `yaml:"clientID"`

	// Enable the configuration store for authenticated users. A folder MUST be set.
	ConfigStoreEnabled bool `yaml:"configStoreEnabled"`

	// Set the config store folder. Required for enabling the config store
	ConfigStoreFolder string `yaml:"configStoreFolder"`

	// PasswordFile to read from. Use "" for default defined in 'unpwstore.DefaultPasswordFile'
	PasswordFile string `yaml:"passwordFile"`
}

// AuthService for handling authentication and token refresh requests
//  1. Handle login requests and issue JWT tokens
//  2. Handle refresh requests and re-issue JWT tokens
//  3. Handle config requests to persist client configuration
//
type AuthService struct {
	config        AuthServiceConfig
	configStore   *configstore.ConfigStore
	running       bool
	tlsServer     *tlsserver.TLSServer
	signingKey    *ecdsa.PrivateKey
	authenticator *authenticate.Authenticator
}

// EnableConfigStore listens for configuration store requests.
// URL is PUT/GET {server}/auth/config/{appID}
//  storeFolder Folder to store user configuration files
func (srv *AuthService) EnableConfigStore(storeFolder string) {
	srv.configStore = configstore.NewConfigStore(storeFolder)
	err := srv.configStore.Open()
	if err != nil {
		logrus.Errorf("Failed opening user configuration store: %s", err)
		return
	}

	srv.tlsServer.AddHandler(tlsclient.DefaultJWTConfigPath+"/{"+appIDParam+"}", srv.ServeConfig)
}

// ServeConfig serves or updates user configuration [GET/PUT]
func (srv *AuthService) ServeConfig(userID string, resp http.ResponseWriter, req *http.Request) {
	logrus.Infof("AuthService.ServeConfig: userID=%s", userID)
	appID := mux.Vars(req)[appIDParam]
	if req.Method == http.MethodPut {
		logrus.Warningf("ServeConfig, updated configuration for user '%s'", userID)
		payload, err := ioutil.ReadAll(req.Body)
		if err != nil {
			srv.tlsServer.WriteBadRequest(resp, "Missing payload")
			return
		}
		_ = srv.configStore.Put(userID, appID, string(payload))
	} else if req.Method == http.MethodGet {
		payload := srv.configStore.Get(userID, appID)
		_, _ = resp.Write([]byte(payload))
		return
	} else {
		logrus.Warningf("ServeConfig, method %s not support for user %s", req.Method, userID)
		srv.tlsServer.WriteBadRequest(resp, "Bad method "+req.Method)
		return
	}
}

// SetPassword for updating a user's password
func (srv *AuthService) SetPassword(userID, password string) error {
	return srv.authenticator.SetPassword(userID, password)
}

// Start listening for login and refresh requests
func (srv *AuthService) Start() error {
	var err error

	// call me as often as you like, or is this an error?
	if srv.running {
		err := fmt.Errorf("AuthService is already running")
		logrus.Error(err)
		return err
	}
	pwStore := unpwstore.NewPasswordFileStore(srv.config.PasswordFile, srv.config.ClientID)
	srv.authenticator = authenticate.NewAuthenticator(pwStore)
	err = srv.authenticator.Start()
	if err != nil {
		return err
	}

	err = srv.tlsServer.Start()
	if err != nil {
		logrus.Errorf("AuthService.Start: Error starting authservice: %s", err)
		return err
	}
	// add auth handlers
	srv.tlsServer.EnableJwtAuth(&srv.signingKey.PublicKey)
	srv.tlsServer.EnableJwtIssuer(srv.signingKey, srv.authenticator.VerifyUsernamePassword)

	if srv.config.ConfigStoreEnabled && srv.config.ConfigStoreFolder != "" {
		srv.EnableConfigStore(srv.config.ConfigStoreFolder)
	}
	srv.running = true
	return nil
}

// Stop the authservice
func (srv *AuthService) Stop() {
	if srv.running {
		srv.running = false
		srv.authenticator.Stop()
		srv.tlsServer.Stop()
	}
}

// NewJwtAuthService creates a new instance of a TLS authservice for JWT authentication.
//
// The signing key contains the key needed for JWT token generation and verification.
// For single signon the server certificate public key can be used for verification if the
// certificates are generated using the same key pair.
//
//  config 		service configuration
//  signingKey  private/public key used to sign and verify JWT tokens. nil to use the server certificate keys.
//  serverCert  server own TLS certificate, signed with ecdsa keys
//  caCert      CA certificate to verify client certificates
//  verifyUsernamePassword function of the password store to verify username/password
// This returns the authentication authservice instance.
func NewJwtAuthService(
	config AuthServiceConfig,
	signingKey *ecdsa.PrivateKey,
	serverCert *tls.Certificate,
	caCert *x509.Certificate) *AuthService {

	if config.Port == 0 {
		config.Port = DefaultAuthServicePort
	}
	if config.ClientID == "" {
		config.ClientID = PluginID
	}

	// The TLS server authenticates a request.
	tlsServer := tlsserver.NewTLSServer(config.Address, config.Port, serverCert, caCert)

	if signingKey == nil {
		signingKey = serverCert.PrivateKey.(*ecdsa.PrivateKey)
	}

	srv := AuthService{
		config:     config,
		tlsServer:  tlsServer,
		signingKey: signingKey,
	}
	return &srv
}
