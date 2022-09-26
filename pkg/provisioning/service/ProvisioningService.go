package service

import (
	"crypto/md5"
	"crypto/x509"
	"fmt"
	"sync"

	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/provisioning"
)

const DefaultIoTCertValidityDays = 14
const ApprovedSecret = "approved"
const DefaultRetrySec = 12 * 3600

// ProvisioningService handles the provisioning requests.
// This implements the IProvisioningService interface.
//
// Implementation of the messaging protocol is provided by the capnp protocol adapter
//
// This verifies requests against the out-of-bound secret and uses the certificate service to
// issue IoT device certificates.
// If no OOB secret is provided, the request is stored and awaits approval by the administrator.
//
// If enabled, a discovery record is published using DNS-SD to allow potential clients to find the
// address and ports of the provisioning server, and optionally additional services.
type ProvisioningService struct {
	certCapability certs.IDeviceCerts // client with capability to create device certificates

	// runtime status
	running    bool
	oobSecrets map[string]provisioning.OOBSecret        // [deviceID]secret simple in-memory store for OOB secrets
	requests   map[string]provisioning.ProvisionRequest // open requests
	mux        sync.Mutex
}

// AddOOBSecret adds one or more OOB Secrets for pre-approval and automatic provisioning
// OOBSecrets are kept in-memory until restart or they expire
func (ps *ProvisioningService) AddOOBSecret(secrets []provisioning.OOBSecret) {
	for _, secret := range secrets {
		ps.oobSecrets[secret.DeviceID] = secret
	}
}

// ApproveRequest approves a pending request
// The next time the request is made, it will be accepted
func (ps *ProvisioningService) ApproveRequest(deviceID string) {
	ps.oobSecrets[deviceID] = provisioning.OOBSecret{
		OobSecret: ApprovedSecret,
	}
}

// GetPendingRequests returns the list of open requests
func (ps *ProvisioningService) GetPendingRequests() []provisioning.ProvisionRequest {
	result := make([]provisioning.ProvisionRequest, 0)
	ps.mux.Lock()
	for _, req := range ps.requests {
		result = append(result, req)
	}
	defer ps.mux.Unlock()
	return result
}

// GetRequestCapability returns the API to issuing provisioning requests
func (ps *ProvisioningService) GetRequestCapability() provisioning.IProvisioningRequest {
	return ps
}

// RefreshDeviceCert returns the API to issuing provisioning requests
func (ps *ProvisioningService) RefreshDeviceCert(
	deviceID string, cert *x509.Certificate) (provResp provisioning.ProvisionResponse, err error) {
	// TODO: verify caller certificate validity
	err = ps.certCapability.IsValid(cert)
	if err == nil {
		// create a new certificate
		pubKeyPEM, err := certsclient.PublicKeyToPEM(cert.PublicKey)
	}
	if err == nil {
		provResp.ClientCertPEM, provResp.CaCertPEM, err = ps.certCapability.CreateDeviceCert(
			deviceID, pubKeyPEM, DefaultIoTCertValidityDays)
		provResp.Approved = true
	}
	return provResp, err
}

// SubmitProvisioningRequest handles provisioning request
func (ps *ProvisioningService) SubmitProvisioningRequest(
	provReq provisioning.ProvisionRequest) (provResp provisioning.ProvisionResponse, err error) {

	var approved = false
	provResp = provisioning.ProvisionResponse{}
	if provReq.DeviceID == "" || provReq.MAC == "" || provReq.PubKeyPEM == "" {
		err = fmt.Errorf("missing required request parameters")
		return
	}
	approved, err = ps.verifyApproval(provReq)

	// if a secret is approved, create a certificate
	if approved {
		provResp.ClientCertPEM, provResp.CaCertPEM, err = ps.certCapability.CreateDeviceCert(
			provReq.DeviceID, provReq.PubKeyPEM, DefaultIoTCertValidityDays)
		provResp.Approved = true
	} else {
		provResp.RetrySec = DefaultRetrySec
	}
	if err == nil {
	}
	return provResp, err
}

// Verify if the provisioning request is approved
func (ps *ProvisioningService) verifyApproval(provReq provisioning.ProvisionRequest) (approved bool, err error) {
	// check for manual approval
	secretToMatch, hasSecret := ps.oobSecrets[provReq.DeviceID]
	if !hasSecret {
		// no known secret, so add the request for manual approval
		approved = false
	} else if secretToMatch.OobSecret == ApprovedSecret {
		// manual approval in place
		approved = true
	} else {
		md5ToMatch := fmt.Sprint(md5.Sum([]byte(secretToMatch.OobSecret)))
		if provReq.SecretMD5 == md5ToMatch {
			approved = true
		} else {
			// not a matching secret, reject the request
			approved = false
			err = fmt.Errorf("secret doesn't match for device %s", provReq.DeviceID)
		}
	}
	return approved, err
}
