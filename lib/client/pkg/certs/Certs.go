package certs

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

// LoadX509CertFromPEM loads the x509 certificate from a PEM file format.
func LoadX509CertFromPEM(pemPath string) (cert *x509.Certificate, err error) {
	pemEncoded, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return nil, err
	}
	return X509CertFromPEM(string(pemEncoded))
}

// LoadTLSCertFromPEM loads the TLS certificate from PEM formatted x509 cert and key files
// This is simply a wrapper around tls.LoadX509KeyPair. Since there is a SaveTLSCertToPEM
// it makes sense to have a LoadTLSCertFromPEM.
// If loading fails, this returns nil as certificate pointer
func LoadTLSCertFromPEM(certPEMPath, keyPEMPath string) (cert *tls.Certificate, err error) {
	tlsCert, err := tls.LoadX509KeyPair(certPEMPath, keyPEMPath)
	if err != nil {
		return nil, err
	}
	return &tlsCert, err
}

// SaveTLSCertToPEM saves the x509 certificate and private key to files in PEM format
func SaveTLSCertToPEM(cert *tls.Certificate, certPEMPath, keyPEMPath string) error {
	b := pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}
	certPEM := pem.EncodeToMemory(&b)
	err := ioutil.WriteFile(certPEMPath, certPEM, 0644)
	if err != nil {
		return err
	}
	err = SaveKeysToPEM(cert.PrivateKey, keyPEMPath)

	return err
}

// SaveX509CertToPEM saves the x509 certificate to file in PEM format.
// Clients that receive a client certificate from provisioning can use this
// to save the provided certificate to file.
func SaveX509CertToPEM(cert *x509.Certificate, pemPath string) error {
	b := pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}
	certPEM := pem.EncodeToMemory(&b)
	err := ioutil.WriteFile(pemPath, certPEM, 0644)
	return err
}

// X509CertFromPEM a X509 certificate in PEM format to an X509 instance
func X509CertFromPEM(certPEM string) (*x509.Certificate, error) {
	caCertBlock, _ := pem.Decode([]byte(certPEM))
	if caCertBlock == nil {
		return nil, errors.New("ConverX509CertFromPEM pem.Decode failed")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	return caCert, err
}

// X509CertToPEM converts the x509 certificate to PEM format.
func X509CertToPEM(cert *x509.Certificate) string {
	b := pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}
	certPEM := pem.EncodeToMemory(&b)
	return string(certPEM)
}
