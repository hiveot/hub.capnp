package listener

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
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

// CreateListener starts a network listener on the given address and port
//
// If noTLS is set then serverCert and caCert can be nil. Obviously the
// listener will not use encryption.
//
// NOTE: if an address is given, then this no longer listens on localhost. Any
//
//	 locally run service or SSR client using localhost will not be able to connect.
//
//		addr IP address. "" to listen on all addresses
//		port to listen on
//		noTLS to listen on the port without encryption
//		serverCert server's TLS certificate to authenticate as, or nil if noTLS is true
//		caCert server's CA to authenticate as, or nil if noTLS is true
func CreateListener(
	addr string, port int, noTLS bool, serverCert *tls.Certificate, caCert *x509.Certificate) (
	lis net.Listener, err error) {

	logrus.Infof("creating listener on %s:%d", addr, port)
	// TODO: listen on multiple interfaces if an address is given???
	lis, err = net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if !noTLS {
		lis = CreateTLSListener(lis, serverCert, caCert)
	}
	return lis, err
}
