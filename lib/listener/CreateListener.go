package listener

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"strings"
)

// CreateTLSListener wraps the given listener in TLS v1.3
func CreateTLSListener(
	lis net.Listener, serverCert *tls.Certificate, caCert *x509.Certificate) net.Listener {

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{*serverCert},
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS13,
		RootCAs:      caCertPool,
		ServerName:   "HiveOT Hub",
	}
	tlsLis := tls.NewListener(lis, &tlsConfig)
	return tlsLis
}

// CreateListener starts a network listener on the given address.
//
// If noTLS is set then serverCert and caCert can be nil. Obviously the
// listener will not use encryption.
//
//	addrPort in the format address:port
//	noTLS to listen on the port without encryption
//	serverCert server's TLS certificate to authenticate as if noTLS is false
//	caCert server's CA to authenticate as if noTLS is false
func CreateListener(
	addrPort string, noTLS bool, serverCert *tls.Certificate, caCert *x509.Certificate) (
	lis net.Listener, err error) {

	// remove any trailing /path in case of WS
	parts := strings.Split(addrPort, "/")

	lis, err = net.Listen("tcp", parts[0])
	if !noTLS {
		lis = CreateTLSListener(lis, serverCert, caCert)
	}
	return lis, err
}
