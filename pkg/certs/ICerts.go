// Package certs with POGS capability definitions of the certificate services.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package certs

import "context"

// Default validity of generated service certificates
const defaultServiceCertValidityDays = 30

// Default validity of generated client certificates
const defaultClientCertValidityDays = 30

// Default validity of generated device certificates
const defaultDeviceCertValidityDays = 30

// ICerts defines a POGS based capability API of the cert service
// This interface aggregates all certificate capabilities.
// This approach is experimental and intended to improve security by providing capabilities based on
// user credentials, enforced by the capnp protocol.
type ICerts interface {

	// CapDeviceCerts provides the capability to manage device certificates
	CapDeviceCerts() IDeviceCerts

	// CapServiceCerts provides the capability to manage service certificates
	CapServiceCerts() IServiceCerts

	// CapUserCerts provides the capability to manage user certificates
	CapUserCerts() IUserCerts

	// CapVerifyCerts provides the capability to verify certificates
	CapVerifyCerts() IVerifyCerts

	// Release the provided capabilities
	Release()
}

// IDeviceCerts defines the POGS based capability to create device certificates
type IDeviceCerts interface {
	// CreateDeviceCert generates or renews IoT device certificate for access hub IoT gateway
	//  deviceID is the unique device's ID
	//  pubkeyPEM is the device's public key in PEM format
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateDeviceCert(
		ctx context.Context, deviceID string, pubKeyPEM string, validityDays int) (
		certPEM string, caCertPEM string, err error)
}

// IServiceCerts defines the POGS based capability to create service certificates
type IServiceCerts interface {
	// CreateServiceCert generates a hub service certificate
	//  serviceID is the unique service ID, for example hostname-serviceName
	//  pubkeyPEM is the device's public key in PEM format
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateServiceCert(
		ctx context.Context, serviceID string, pubKeyPEM string, names []string, validityDays int) (
		certPEM string, caCertPEM string, err error)
}

// IUserCerts defines the POGS based capability to create user certificates
type IUserCerts interface {

	// CreateUserCert generates an end-user certificate for access hub gateway services
	// Intended for users that use certificates instead of regular login.
	//  userID is the unique user's ID, for example an email address
	//  pubkeyPEM is the user's public key in PEM format
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateUserCert(
		ctx context.Context, userID string, pubKeyPEM string, validityDays int) (
		certPEM string, caCertPEM string, err error)
}

// IVerifyCerts defines the POGS based capability to verify an issued certificate
type IVerifyCerts interface {
	// VerifyCert verifies if the certificate is valid for the Hub
	VerifyCert(ctx context.Context, clientID string, certPEM string) error
}
