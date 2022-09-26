// Package certs with POGS capability definitions of the certificate services.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package certs

// Default validity of generated service certificates
const defaultServiceCertValidityDays = 30

// Default validity of generated client certificates
const defaultClientCertValidityDays = 30

// Default validity of generated device certificates
const defaultDeviceCertValidityDays = 30

// ICerts defines a POGS based capability API of the cert service
// This interface aggregates all certificate capabilities.
// This approach is experimental and intended to separate capabilities using the capnp protocol.
type ICerts interface {
	IDeviceCerts
	IVerifyCert
	IServiceCerts
	IUserCerts
}

// IDeviceCerts defines the POGS based capability to create device certificates
type IDeviceCerts interface {
	// CreateDeviceCert generates or renews IoT device certificate for access hub IoT gateway
	//  deviceID is the unique device's ID
	//  pubkeyPEM is the device's public key in PEM format
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateDeviceCert(deviceID string, pubKeyPEM string, validityDays int) (
		certPEM string, caCertPEM string, err error)
}

// IServiceCerts defines the POGS based capability to create service certificates
type IServiceCerts interface {
	// CreateServiceCert generates a hub service certificate
	//  serviceID is the unique service ID, for example hostname-serviceName
	//  pubkeyPEM is the device's public key in PEM format
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateServiceCert(serviceID string, pubKeyPEM string, names []string, validityDays int) (
		certPEM string, caCertPEM string, err error)
}

// IUserCerts defines the POGS based capability to create user certificates
type IUserCerts interface {

	// CreateUserCert generates an end-user certificate for access hub gateway services
	// Intended for users that use certificates instead of regular login.
	//  userID is the unique user's ID, for example an email address
	//  pubkeyPEM is the user's public key in PEM format
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateUserCert(userID string, pubKeyPEM string, validityDays int) (
		certPEM string, caCertPEM string, err error)
}

// IVerifyCert defines the POGS based capability to verify an issued certificate
type IVerifyCert interface {
	// VerifyCert verifies if the certificate is valid for the Hub and not revoked
	VerifyCert(clientID string, certPEM string) error
}
