package provisioning

import "crypto/x509"

// OOBSecret holds a device's Out Of Band secret for automated provisioning
// If the deviceID and MD5 hash of the secret match with the request it will be approved immediately
type OOBSecret struct {

	// The unique device ID or MAC address
	DeviceID string

	// The OOB secret of the device, "approved" to accept any secret
	OobSecret string
}

// ProvisionRequest holds the data of a provisioning request
type ProvisionRequest struct {

	// deviceID is the required unique ID of the device on the local network.
	DeviceID string

	// mac is required IoT device/service MAC address
	MAC string

	// secretMD5 is the optional MD5 hash of the out-of-band secret if available. Use "" for manual approval
	SecretMD5 string

	// pubKeyPEM is the required public key of the IoT device, used to generate the client certificate.
	PubKeyPEM string

	// ISO8601 optional timestamp until which the provisioning request is valid. Default is 1 hour.
	//ValidUntil string

	// ISO8601 timestamp when the request was received. This is set by the service.
	Timestamp string
}

// ProvisionResponse holds the response to a provisioning request
type ProvisionResponse struct {

	// the request is approved, if false, wait retrySec to retry
	Approved bool

	// The issued client certificate if approved, in PEM format
	ClientCertPEM string

	// CA's certificate used to sign the request, in PEM format
	CaCertPEM string

	// Optional delay for retrying the request in seconds in case status is pending
	RetrySec int
}

// IProvisioningService defines a POGS based interface of the provisioning service
// Intended for administrators
type IProvisioningService interface {

	// AddOOBSecret adds a list of OOB secrets
	AddOOBSecret(oobSecrets []OOBSecret)

	// ApproveRequest approves a pending request for the given device ID
	ApproveRequest(deviceID string)

	// GetPendingRequests returns a list of pending requests
	GetPendingRequests() []ProvisionRequest

	// GetRequestCapability returns the capability to request provisioning
	GetRequestCapability() IProvisioningRequest
}

// IProvisioningRequest defines the capability to request or refresh a provisioning certificate
// Intended for use by IoT devices.
type IProvisioningRequest interface {
	// RefreshDeviceCert refreshes a device certificate with a new expiry date
	//
	// This will only succeed if the request is made with a valid certificate and this
	// certificate.
	//
	//  deviceID of the device to request the certificate for.
	//  certPEM the current certificate in PEM format
	RefreshDeviceCert(deviceID string, cert *x509.Certificate) (ProvisionResponse, error)

	// SubmitProvisioningRequest handles the provisioning request.
	// If the deviceID and MD5 hash of the secret match with previously uploaded secrets then the
	// request will be approved immediately.
	// This returns an error if the request is invalid
	SubmitProvisioningRequest(provRequest ProvisionRequest) (ProvisionResponse, error)
}
