package config

import (
	"crypto/ecdsa"
	"crypto/x509"
)

// Default certificate and private key file names
const (
	DefaultCaCertFile     = "caCert.pem"
	DefaultCaKeyFile      = "caKey.pem"
	DefaultPluginCertFile = "pluginCert.pem"
	DefaultPluginKeyFile  = "pluginKey.pem"
	DefaultServerCertFile = "serverCert.pem"
	DefaultServerKeyFile  = "serverKey.pem"
	DefaultAdminCertFile  = "adminCert.pem"
	DefaultAdminKeyFile   = "adminKey.pem"
)
const DefaultServiceCertDurationDays = 7
const DefaultClientCertDurationDays = 7
const DefaultDeviceCertDurationDays = 7

// ServiceName is the name of this service for logging
const ServiceName = "certsvc"

// CertSvcConfig with configuration for the certificate service
type CertSvcConfig struct {
	CaCert *x509.Certificate
	CaKey  *ecdsa.PrivateKey
}

// NewCertSvcConfig creates a new config with default values
func NewCertSvcConfig(caCert *x509.Certificate, caKey *ecdsa.PrivateKey) CertSvcConfig {
	cfg := CertSvcConfig{
		CaCert: caCert,
		CaKey:  caKey,
	}
	return cfg
}
