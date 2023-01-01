package listener

import (
	"crypto/tls"
	"crypto/x509"
	"net"
)

// CreateTLSListener wraps the given listener in TLS v1.3
// TODO: verify client certificates
func CreateTLSListener(
	lis net.Listener, serverCert *tls.Certificate, caCert *x509.Certificate) net.Listener {

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	tlsConfig := tls.Config{
		Certificates:       []tls.Certificate{*serverCert},
		ClientAuth:         tls.VerifyClientCertIfGiven,
		ClientCAs:          caCertPool,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
		RootCAs:            caCertPool,
		ServerName:         "HiveOT Hub",
	}
	tlsLis := tls.NewListener(lis, &tlsConfig)
	return tlsLis
}
