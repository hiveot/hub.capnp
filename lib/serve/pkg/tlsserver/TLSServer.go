// Package tlsserver with TLS server for use by plugins and testing
package tlsserver

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/rs/cors"
	"net/http"
	"strings"
	"time"

	"github.com/wostzone/hub/lib/client/pkg/tlsclient"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// TLSServer is a simple TLS Server supporting BASIC, Jwt and client certificate authentication
type TLSServer struct {
	address           string
	port              uint
	caCert            *x509.Certificate
	serverCert        *tls.Certificate
	httpServer        *http.Server
	router            *mux.Router
	httpAuthenticator *HttpAuthenticator

	jwtIssuer *JWTIssuer
}

// AddHandler adds a new handler for a path.
//
// The server authenticates the request before passing it to this handler.
// The handler's userID is that of the authenticated user, and is intended for authorization of the request.
// If authentication is not enabled then the userID is empty.
//
// apply .Method(http.MethodXyz) to restrict the accepted HTTP methods
//
//  path to listen on. This supports wildcards
//  handler to invoke with the request. The userID is only provided when an authenticator is used
// Returns the route. Apply '.Method(http.MethodPut|Post|Get)' to restrict the accepted HTTP methods
func (srv *TLSServer) AddHandler(path string,
	handler func(userID string, resp http.ResponseWriter, req *http.Request)) *mux.Route {

	// do we need a local copy of handler? not sure
	local_handler := handler

	// the internal authenticator performs certificate based, basic or jwt token authentication if needed
	route := srv.router.HandleFunc(path, func(resp http.ResponseWriter, req *http.Request) {
		// test, allow CORS if enabled.
		if req.Method == http.MethodOptions {
			// don't return a payload with the cors options request
			return
		}

		// valid authentication without userID means a plugin certificate was used which is always authorized
		userID, match := srv.httpAuthenticator.AuthenticateRequest(resp, req)
		if !match {
			msg := fmt.Sprintf("TLSServer.HandleFunc %s: User '%s' from %s is unauthorized", path, userID, req.RemoteAddr)
			logrus.Infof("%s", msg)
			srv.WriteForbidden(resp, msg)
		} else {
			local_handler(userID, resp, req)
		}
	})
	return route
}

// AddHandlerNoAuth adds a new handler for a path that does not require authentication
// The server passes the request directly to the handler
//
//  path to listen on. This supports wildcards
//  handler to invoke with the request. The userID is only provided when an authenticator is used
// Returns the route. Apply '.Method(http.MethodPut|Post|Get)' to restrict the accepted HTTP methods
func (srv *TLSServer) AddHandlerNoAuth(path string,
	handler func(resp http.ResponseWriter, req *http.Request)) *mux.Route {

	route := srv.router.HandleFunc(path, func(resp http.ResponseWriter, req *http.Request) {
		handler(resp, req)
	})
	return route

}

// EnableBasicAuth enables BASIC authentication on this server
// Basic auth is a legacy authentication scheme and not recommended as it requires each service to
// have access to the credentials store. Use of JwtAuth is preferred.
//
// validateCredentials is the function that verifies the given credentials
func (srv *TLSServer) EnableBasicAuth(validateCredentials func(loginName string, password string) bool) {
	srv.httpAuthenticator.EnableBasicAuth(validateCredentials)
}

// EnableJwtAuth enables JWT authentication using asymmetric keys
// JWT tokens are included in the head authorization field and signed by an issuing authentication server using
// the server's private key. The provided verification key is the server's public key.
//  verificationKey is the public key to verify tokens. Use nil to use the server own public key
func (srv *TLSServer) EnableJwtAuth(verificationKey *ecdsa.PublicKey) {
	if verificationKey == nil {
		issuerKey := srv.serverCert.PrivateKey.(*ecdsa.PrivateKey)
		verificationKey = &issuerKey.PublicKey
	}
	srv.httpAuthenticator.EnableJwtAuth(verificationKey)
}

// EnableJwtIssuer enables JWT token issuer using asymmetric keys
// Token are issued using the PUT /login request with payload carrying {username: , password:}
// Tokens are refreshed using the PUT /refresh request
//
// The login/refresh paths are defined in tlsclient.DefaultJWTLoginPath, tlsclient.DefaultJWTRefreshPath
//
// issuerKey is the private key used to sign the tokens. Use nil to use the server's private key
// validateCredentials is the handler that matches credentials with those in the credentials store
func (srv *TLSServer) EnableJwtIssuer(issuerKey *ecdsa.PrivateKey,
	validateCredentials func(loginName string, password string) bool) {
	// for now the JWT login/refresh paths are fixed. Once a use-case comes up that requires something configurable
	// this can be updated.
	jwtLoginPath := tlsclient.DefaultJWTLoginPath
	hwtRefreshPath := tlsclient.DefaultJWTRefreshPath
	if issuerKey == nil {
		issuerKey = srv.serverCert.PrivateKey.(*ecdsa.PrivateKey)
	}
	// handler of issuing JWT tokens
	srv.jwtIssuer = NewJWTIssuer("tlsserver", issuerKey, validateCredentials)
	srv.AddHandlerNoAuth(jwtLoginPath, srv.jwtIssuer.HandleJWTLogin).Methods(http.MethodPost, http.MethodOptions)
	srv.AddHandlerNoAuth(hwtRefreshPath, srv.jwtIssuer.HandleJWTRefresh).Methods(http.MethodPost, http.MethodOptions)

}

// Start the TLS server using the provided CA and Server certificates.
// If a client certificate is provided it must be valid.
// This configures handling of CORS requests to allow:
//  - any origin by returning the requested origin (not using wildcard '*').
//  - any method, eg PUT, POST, GET, PATCH,
//  - headers "Origin", "Accept", "Content-Type", "X-Requested-With"
func (srv *TLSServer) Start() error {
	var err error

	logrus.Infof("TLSServer.Start Starting TLS server on address: %s:%d.", srv.address, srv.port)
	if srv.caCert == nil || srv.serverCert == nil {
		err := fmt.Errorf("TLSServer.Start: missing CA or server certificate")
		logrus.Error(err)
		return err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(srv.caCert)

	serverTLSConf := &tls.Config{
		Certificates:       []tls.Certificate{*srv.serverCert},
		ClientAuth:         tls.VerifyClientCertIfGiven,
		ClientCAs:          caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	// handle CORS using the cors plugin
	// see also: https://stackoverflow.com/questions/43871637/no-access-control-allow-origin-header-is-present-on-the-requested-resource-whe
	// TODO: add configuration for CORS origin: allowed, sameaddress, exact
	c := cors.New(cors.Options{
		// return the origin as allowed origin
		AllowOriginFunc:  func(orig string) bool {
			// local requests are always allowed, even over http (for testing) - todo: disable in production
			if strings.HasPrefix(orig, "https://127.0.0.1") || strings.HasPrefix(orig, "https://localhost") ||
				strings.HasPrefix(orig, "http://127.0.0.1") || strings.HasPrefix(orig, "http://localhost") {
				return true
			} else if strings.HasPrefix(orig, "https://"+srv.address) {
				return true
			}
			return false
		},
		// default allowed headers is "Origin", "Accept", "Content-Type", "X-Requested-With" (missing authorization)
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "Authorization"},
		// default is get/put/patch/post/delete/head
		//AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch},
		Debug:            true,
		AllowCredentials: true,
	})
	handler := c.Handler(srv.router)

	srv.httpServer = &http.Server{
		Addr: fmt.Sprintf("%s:%d", srv.address, srv.port),
		// ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
		// WriteTimeout: 10 * time.Second,
		Handler:   handler,
		TLSConfig: serverTLSConf,
	}
	// mutex to capture error result in case startup in the background failed
	go func() {

		// serverTLSConf contains certificate and key
		err2 := srv.httpServer.ListenAndServeTLS("", "")
		if err2 != nil && err2 != http.ErrServerClosed {
			err = fmt.Errorf("TLSServer.Start: ListenAndServeTLS: %s", err2)
			logrus.Error(err)
		}
	}()
	// Make sure the server is listening before continuing
	time.Sleep(time.Second)
	return err
}

// Stop the TLS server and close all connections
func (srv *TLSServer) Stop() {
	logrus.Infof("TLSServer.Stop: Stopping TLS server")

	if srv.httpServer != nil {
		srv.httpServer.Shutdown(context.Background())
	}
}

// NewTLSServer creates a new TLS Server instance with authentication support.
// Use AddHandler to handle incoming requests for the given route and indicate if authentication is required.
//
// The following authentication methods are supported:
//  Certificate based auth using the caCert to verify client certificates
//  Basic authentication if 'EnableBasicAuth' is used.
//  JWT asymmetric token authentication if EnableJwtAuth is used.
//
//  address        server listening address
//  port           listening port
//  serverCert     Server TLS certificate
//  caCert         CA certificate to verify client certificates
//
// returns TLS server for handling requests
func NewTLSServer(address string, port uint,
	serverCert *tls.Certificate,
	caCert *x509.Certificate,
) *TLSServer {

	srv := &TLSServer{
		caCert:     caCert,
		serverCert: serverCert,
		router: mux.NewRouter(),
	}
	//// support for CORS response headers
	//srv.router.Use(mux.CORSMethodMiddleware(srv.router))

	//issuerKey := serverCert.PrivateKey.(*ecdsa.PrivateKey)
	//serverX509, _ := x509.ParseCertificate(serverCert.Certificate[0])
	//pubKey := certs.PublicKeyFromCert(serverX509)

	// Authenticate incoming https requests
	srv.httpAuthenticator = NewHttpAuthenticator()

	srv.address = address
	srv.port = port
	return srv
}
