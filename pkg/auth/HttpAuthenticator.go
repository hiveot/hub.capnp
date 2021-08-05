package tlsserver

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/pkg/certsetup"
)

// Connection headers
const (
	AuthorizationHeader = "Authorization"
	ClientHeader        = "Client"
)

// HTTP auth for authenticating HTTP requests
// This handles certificate, basic and digest(*) authentication.
//
// Digest auth only works for matching WoST hashes - argon2id - which is not an offical digest hash
type HttpAuthenticator struct {

}
// move this to the hub auth package

// Test if the request is made using a valid client certificate
func (hauth *HttpAuthenticator) IsValidCert(request *http.Request) bool {
	if len(request.TLS.PeerCertificates) == 0 {
		return false
	}

	// plugins and admins have full permission
	cert := request.TLS.PeerCertificates[0]
	certOU := cert.Issuer.OrganizationalUnit
	if certOU == certsetup.OUPlugin || certOU == certsetup.OUAdmin {
		return true
	}
	if ou == OU
	return cert
}

// Authentication of HTTP requests
// DIGEST auth must use the validator's hashing algorithm. The default for
// WoST is argon2id but this can change
//  - If the caller has a valid client certificate, accept the request
//  - If the caller uses BASIC or DIGEST Auth then validate the given secret
func (hauth *HttpAuthenticator) Authenticate(
	validator func(loginID string, algo string, secret string) bool,
	request *http.Request) bool {

	clientID := request.Header.Get(ClientHeader)

	if clientID == "" {
		// http.Error(response, "Invalid client. A clientID is required.", 401)
		logrus.Warningf("Missing clientID from client '%s'", request.RemoteAddr)
		return false
	}

	logrus.Infof("Incoming connection from %s", clientID)
	return true
}

// Create a new authenticator instance for the given handler
// The authenticator validates a new incoming connection using basic, digest or client certificate
// based authentication.
//
//  validator is the function used to validate a secret for a loginID. The secret is typically
// the hash of the password.
//  handler is the handler to invoke if authentication is accepted by the validator
func (hauth *HttpAuthenticator) NewAuthenticator(
	validator func(loginID string, secret string) bool,
	handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(resp http.ResponseWriter, req *http.Request) {
		if srv.Authenticate(validator, req) {
			handler(resp, req)
		} else {
			srv.WriteUnauthorized(resp, "Invalid username or password")
		}
	}
}
