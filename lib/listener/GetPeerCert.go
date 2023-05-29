package listener

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
)

// GetPeerCert returns a peer certificate info from a tls connection along
// with the client's ID and auth type in the commonName and OU fields.
// returns nil if no peer certificate is provided
func GetPeerCert(conn net.Conn) (peerCert *x509.Certificate, clientID string, ou string, err error) {
	ou = "" // unauthenticated
	tlsCon, isValid := conn.(*tls.Conn)
	if !isValid {
		err = errors.New("connection is not a TLS connection")
		return
	}
	err = tlsCon.Handshake()
	// not a valid TLS connection so drop it
	if err != nil {
		err = fmt.Errorf("dropping invalid TLS connection from '%s':%w", tlsCon.RemoteAddr(), err)
		_ = conn.Close()
		return
	}

	cstate := tlsCon.ConnectionState()
	if len(cstate.PeerCertificates) == 0 {
		err = errors.New("connection is made without peer certificate")
		return
	}
	peerCert = cstate.PeerCertificates[0]
	clientID = peerCert.Subject.CommonName
	if len(peerCert.Subject.OrganizationalUnit) > 0 {
		ou = peerCert.Subject.OrganizationalUnit[0]
	}
	return peerCert, clientID, ou, nil
}
