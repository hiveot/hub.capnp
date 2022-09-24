// Package client with POGS definitions of the cert service.
// Unfortunately capnp does generate POGS types so we need to duplicate them
package client

// Default validity of generated service certificates
const defaultServiceCertValidityDays = 30

// Default validity of generated client certificates
const defaultClientCertValidityDays = 30

// Default validity of generated device certificates
const defaultDeviceCertValidityDays = 30

// ICertService defines a POGS based capability API of the cert service
// This is implemented by the service itself and by the client wrapper.
// Capnp (or other) RPC adapters copy between pogs and their internal format.
type ICertService interface {
	// CreateDeviceCert generates or renews IoT device certificate for access hub IoT gateway
	//  deviceID is the unique device's ID
	//  pubkeyPEM is the device's public key in PEM format
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateDeviceCert(deviceID string, pubKeyPEM string, validityDays int) (
		certPEM string, caCertPEM string, err error)

	// CreateServiceCert generates a hub service certificate
	//  serviceID is the unique service ID, for example hostname-serviceName
	//  pubkeyPEM is the device's public key in PEM format
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateServiceCert(serviceID string, pubKeyPEM string, names []string, validityDays int) (
		certPEM string, caCertPEM string, err error)

	// CreateUserCert generates an end-user certificate for access hub gateway services
	// Intended for users that use certificates instead of regular login.
	//  userID is the unique user's ID, for example an email address
	//  pubkeyPEM is the user's public key in PEM format
	//  validityDays is the duration the cert is valid for. Use 0 for default.
	CreateUserCert(userID string, pubKeyPEM string, validityDays int) (
		certPEM string, caCertPEM string, err error)
}
