// Package certs with POGS capability definitions of the certificate services.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package certs

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

const DefaultCACertName = hubapi.DefaultCaCertFile

// DefaultServiceCertValidityDays with validity of generated service certificates
const DefaultServiceCertValidityDays = int(hubapi.DefaultServiceCertValidityDays)

// DefaultUserCertValidityDays with validity of generated client certificates
const DefaultUserCertValidityDays = int(hubapi.DefaultUserCertValidityDays)

// DefaultDeviceCertValidityDays with validity of generated device certificates
const DefaultDeviceCertValidityDays = int(hubapi.DefaultDeviceCertValidityDays)

// ServiceName to connect to the service
const ServiceName = "certs"

// ICerts defines a POGS based capability API of the cert service
// This interface aggregates all certificate capabilities.
// This approach is experimental and intended to improve security by providing capabilities based on
// user credentials, enforced by the capnp protocol.
type ICerts interface {

	// CapDeviceCerts provides the capability to manage device certificates
	CapDeviceCerts(ctx context.Context, clientID string) IDeviceCerts

	// CapServiceCerts provides the capability to manage service certificates
	CapServiceCerts(ctx context.Context, clientID string) IServiceCerts

	// CapUserCerts provides the capability to manage user certificates
	CapUserCerts(ctx context.Context, clientID string) IUserCerts

	// CapVerifyCerts provides the capability to verify certificates
	CapVerifyCerts(ctx context.Context, clientID string) IVerifyCerts
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

	// Release the capability and its resources after use
	Release()
}

// IServiceCerts defines the POGS based capability to create service certificates
type IServiceCerts interface {
	// CreateServiceCert generates a hub service certificate
	// This returns the PEM encoded certificate with certificate of the CA that signed it.
	// An error is returned if one of the parameters is invalid.
	//
	//  serviceID is the unique service ID used as the CN. for example hostname-serviceName
	//  pubkeyPEM is the device's public key in PEM format
	//  names are the SAN names to include with the certificate, typically the service IP address or host names
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateServiceCert(
		ctx context.Context, serviceID string, pubKeyPEM string, names []string, validityDays int) (
		certPEM string, caCertPEM string, err error)

	// Release the capability and its resources after use
	Release()
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

	// Release the capability and its resources after use
	Release()
}

// IVerifyCerts defines the POGS based capability to verify an issued certificate
type IVerifyCerts interface {
	// VerifyCert verifies if the certificate is valid for the Hub
	VerifyCert(ctx context.Context, clientID string, certPEM string) error

	// Release the capability and its resources after use
	Release()
}
