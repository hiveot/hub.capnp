package tlsserver

import (
	"net/http"
)

// CertAuthenticator verifies the client certificate authentication is used
// This simply checks if a client certificate is active and assumes that having one is sufficient to pass auth
type CertAuthenticator struct {
}

// AuthenticateRequest
// The real check happens by the TLS server that verifies it is signed by the CA.
// If the certificate is a plugin, then no userID is returned
// Returns the userID of the certificate (CN) or an error if no client certificate is used
func (hauth *CertAuthenticator) AuthenticateRequest(resp http.ResponseWriter, req *http.Request) (userID string, ok bool) {
	if len(req.TLS.PeerCertificates) == 0 {
		return "", false
	}
	cert := req.TLS.PeerCertificates[0]
	userID = cert.Subject.CommonName
	// a plugin is not a username
	if cert.Subject.CommonName == "plugin" {
		userID = ""
	}

	return userID, true
}

// NewCertAuthenticator creates a new HTTP authenticator
// Use .AuthenticateRequest() to authenticate the incoming request
func NewCertAuthenticator() *CertAuthenticator {
	ca := &CertAuthenticator{}
	return ca
}
