// Package idprovclient with IDProv protocol message definitions
package idprovclient

// IDProvDirectoryPath contains the default request path to get the IDProv server directory
const IDProvDirectoryPath = "/idprov/directory"

// Device provisioning status
const (
	ProvisionStatusApproved = "Approved" // request approved. Certificate is available
	ProvisionStatusRejected = "Rejected" // request is rejected
	ProvisionStatusWaiting  = "Waiting"  // waiting for OOB confirmation
)

// DirectoryEndpoints hold the URLs of endpoints
type DirectoryEndpoints struct {
	GetDirectory            string `json:"directory"`
	GetDeviceStatus         string `json:"status"`
	PostOobSecret           string `json:"postOobSecret"`
	PostProvisioningRequest string `json:"postProvisionRequest"`
}

// GetDirectoryMessage contains the API directory from the IDProv server
// This defines the paths, variable names of var
type GetDirectoryMessage struct {
	// URLs of provisioning endpoints
	Endpoints DirectoryEndpoints `json:"endpoints"`
	// Server CA certificate
	CaCertPEM []byte `json:"caCert"`
	// Server version
	Version string `json:"version"`
	// list of services the certificate supports including the message bus, directory server
	Services map[string]string `json:"services"`
}

// GetDeviceStatusMessage contains a device's provisioning status
// This is also the response to a provisioning request.
type GetDeviceStatusMessage struct {
	DeviceID      string `json:"deviceID"`
	Status        string `json:"status"`
	CaCertPEM     string `json:"caCert"`
	ClientCertPEM []byte `json:"clientCert"`
}

// PostOobSecretMessage contains the out of band secret to send to the IDProv server
type PostOobSecretMessage struct {
	DeviceID   string `json:"deviceID"`
	Secret     string `json:"oobSecret"`
	ValidUntil string `json:"validUntil"`
}

// PostProvisionRequestMessage contains the request for device provisioning
type PostProvisionRequestMessage struct {
	DeviceID     string `json:"deviceID"`
	IP           string `json:"ip"`
	MAC          string `json:"mac"`
	PublicKeyPEM string `json:"publicKeyPEM"`
	Signature    string `json:"signature"`
}

// PostProvisionResponseMessage contains the result of the provisioning request
type PostProvisionResponseMessage struct {
	RetrySec      uint   `json:"retrySec,omitempty"`
	Status        string `json:"status"`
	CaCertPEM     string `json:"caCert"`
	ClientCertPEM string `json:"clientCert"`
	Signature     string `json:"signature"`
}
