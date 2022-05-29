// Package authservice to serve a REST api for authentication and token refresh
package authservice

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/wostzone/wost-go/pkg/tlsserver"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/authn/pkg/clientconfigstore"
	"github.com/wostzone/hub/authn/pkg/jwtissuer"
	"github.com/wostzone/hub/authn/pkg/unpwauth"
	"github.com/wostzone/hub/authn/pkg/unpwstore"
	"github.com/wostzone/wost-go/pkg/tlsclient"
)

// PluginID of the service
const PluginID = "authservice"

// DefaultAuthServicePort to connect to the authn service
const DefaultAuthServicePort = 8881

// DefaultAccessTokenValiditySec with access token validity in seconds
const DefaultAccessTokenValiditySec = 3600

// DefaultRefreshTokenValiditySec with Refresh token validity before refresh
const DefaultRefreshTokenValiditySec = 1209600

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

	// Enable the configuration store for authenticated users. Default is true
	ConfigStoreEnabled bool `yaml:"configStoreEnabled"`

	// Set the client config store folder. Default is 'clientconfig' in the config folder
	ConfigStoreFolder string `yaml:"configStoreFolder"`

	// PasswordFile to read from. Use "" for default defined in 'unpwstore.DefaultPasswordFile'
	PasswordFile string `yaml:"passwordFile"`

	// Access token validity. Default is 1 hour
	AccessTokenValiditySec int `yaml:"accessTokenValiditySec"`

	// Refresh token validity. Default is 1209600 (14 days)
	RefreshTokenValiditySec int `yaml:"refreshTokenValiditySec"`
}

// AuthService for handling authentication and token refresh requests
//  1. Handle login requests and issue JWT tokens
//  2. Handle refresh requests and re-issue JWT tokens
//  3. Handle config requests to persist client configuration
//
type AuthService struct {
	config        AuthServiceConfig
	configStore   *clientconfigstore.ClientConfigStore
	running       bool
	tlsServer     *tlsserver.TLSServer
	signingKey    *ecdsa.PrivateKey
	authenticator *unpwauth.UnpwAuthenticator
	jwtIssuer     *jwtissuer.JWTIssuer
}

// EnableConfigStore listens for configuration store requests.
// URL is PUT/GET {server}/authn/config/{appID}
//  storeFolder Folder to store user configuration files
func (srv *AuthService) EnableConfigStore(storeFolder string) {
	srv.configStore = clientconfigstore.NewClientConfigStore(storeFolder)
	err := srv.configStore.Open()
	if err != nil {
		logrus.Errorf("Failed opening user configuration store: %s", err)
		return
	}

	srv.tlsServer.AddHandler(tlsclient.DefaultJWTConfigPath+"/{"+appIDParam+"}", srv.ServeConfig)
}

// EnableJwtIssuer enables JWT token issuer using asymmetric keys
// Token are issued using the PUT /login request with payload carrying {username: , password:}
// Tokens are refreshed using the PUT /refresh request
//
// The login/refresh paths are defined in tlsclient.DefaultJWTLoginPath, tlsclient.DefaultJWTRefreshPath
//
// issuerKey is the private key used to sign the tokens. Use nil to use the server's private key
// validateCredentials is the handler that matches credentials with those in the credentials store
func (srv *AuthService) EnableJwtIssuer(issuerKey *ecdsa.PrivateKey,
	accessTokenValiditySec int, refreshTokenValiditySec int,
	validateCredentials func(loginName string, password string) bool,
) {
	// for now the JWT login/refresh paths are fixed. Once a use-case comes up that requires something configurable
	// this can be updated.
	jwtLoginPath := tlsclient.DefaultJWTLoginPath
	hwtRefreshPath := tlsclient.DefaultJWTRefreshPath
	if issuerKey == nil {
		//issuerKey = srv.serverCert.PrivateKey.(*ecdsa.PrivateKey)
		issuerKey = srv.signingKey
	}
	// handler of issuing JWT tokens
	srv.jwtIssuer = jwtissuer.NewJWTIssuer("AuthService",
		issuerKey,
		accessTokenValiditySec, refreshTokenValiditySec,
		validateCredentials,
	)
	srv.tlsServer.AddHandlerNoAuth(jwtLoginPath, srv.jwtIssuer.HandleJWTLogin).Methods(http.MethodPost, http.MethodOptions)
	srv.tlsServer.AddHandlerNoAuth(hwtRefreshPath, srv.jwtIssuer.HandleJWTRefresh).Methods(http.MethodPost, http.MethodOptions)

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
	srv.authenticator = unpwauth.NewUnPwAuthenticator(pwStore)
	err = srv.authenticator.Start()
	if err != nil {
		return err
	}

	err = srv.tlsServer.Start()
	if err != nil {
		logrus.Errorf("AuthService.Start: Error starting authservice: %s", err)
		return err
	}
	// add authn handlers
	srv.tlsServer.EnableJwtAuth(&srv.signingKey.PublicKey)
	srv.EnableJwtIssuer(
		srv.signingKey,
		srv.config.AccessTokenValiditySec,
		srv.config.RefreshTokenValiditySec,
		srv.authenticator.VerifyUsernamePassword,
	)

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
	if config.AccessTokenValiditySec <= 0 {
		config.AccessTokenValiditySec = DefaultAccessTokenValiditySec
	}
	//logrus.Warningf("NewJwtAuthService: refreshtoken validity from config=%d", config.RefreshTokenValiditySec)
	if config.RefreshTokenValiditySec <= 0 {
		config.RefreshTokenValiditySec = DefaultRefreshTokenValiditySec
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
