package idprovserver

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/certs/pkg/certsetup"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/wost-go/pkg/certsclient"
)

// ServeProvisionRequest handles request to provide a client certificate
//
// A valid Out-of-band secret must be provided if the client does not provide a valid client certificate
//
// If successful a copy of the client certificate is saved in the certstore folder as {deviceID}Cert.pem
// response is:
//   200 OK if request is valid
//   400 if the request is incomplete or missing fields
func (srv *IDProvServer) ServeProvisionRequest(resp http.ResponseWriter, req *http.Request) {
	var provReq idprovclient.PostProvisionRequestMessage
	var validCert = false
	var peerCert *x509.Certificate
	var mySecret = "" // only used if a signature must be verified

	body, err := ioutil.ReadAll(req.Body)
	if err == nil {
		err = json.Unmarshal(body, &provReq)
	}
	if err != nil {
		srv.tlsServer.WriteBadRequest(resp, fmt.Sprintf("ServeProvisionRequest: Invalid body in message: %s", err))
		return
	} else if provReq.DeviceID == "" || provReq.IP == "" || provReq.PublicKeyPEM == "" {
		srv.tlsServer.WriteBadRequest(resp, "ServeProvisionRequest: Missing device info, deviceID, IP or public key")
		return
	}

	// If the client connect with a valid certificate then no need to verify the signature using a secret
	// a Thing OU certificate must be valid and have the same deviceID
	// a Plugin or Admin OU certificate can issue a certificate without an existing deviceID
	for _, peerCert = range req.TLS.PeerCertificates {
		err = srv.validateCertificate(peerCert, provReq.DeviceID)
		if err == nil {
			validCert = true
			break
		}
	}

	// If client doesn't have a valid certificate then a valid signature must be provided,
	// otherwise it is ignored.
	mySecret = srv.oobSecrets[provReq.DeviceID]
	if !validCert {
		// if the secret is unknown then wait for an update of the secret by admin
		if mySecret == "" {
			respStatus := idprovclient.PostProvisionResponseMessage{
				Status: idprovclient.ProvisionStatusWaiting,
			}
			statusMsg, _ := json.Marshal(respStatus)
			logrus.Infof("ServeProvisionRequest: deviceID='%s'. Waiting for OOB secret", provReq.DeviceID)
			resp.Write(statusMsg)
			return
		}
		// Verify the request signature for the message using the OOB secret we received from admin
		copyOfRequest := provReq
		copyOfRequest.Signature = ""
		serialized, _ := json.Marshal(copyOfRequest)
		err = idprovclient.Verify(string(serialized), mySecret, provReq.Signature)
		if err != nil {
			// invalid secret response is 200 with rejected status
			logrus.Infof("ServeProvisionRequest: deviceID='%s'. Invalid OOB secret provided. Request denied", provReq.DeviceID)
			// rejected
			respStatus := idprovclient.PostProvisionResponseMessage{
				Status: idprovclient.ProvisionStatusRejected,
			}
			statusMsg, _ := json.Marshal(respStatus)
			resp.Write(statusMsg)
			return
		}
		logrus.Infof("ServeProvisionRequest: Client '%s' has provided a valid OOB secret. New certificate for device authorized.", provReq.DeviceID)
	}

	// generate or refresh the certificate
	ownerPubKey, err := certsclient.PublicKeyFromPEM(provReq.PublicKeyPEM)
	if err != nil {
		srv.tlsServer.WriteBadRequest(resp, fmt.Sprintf("ServeProvisionRequest: Invalid public key in request for %s: %s", provReq.DeviceID, err))
		return
	}
	newClientCert, err := certsetup.CreateHubClientCert(
		provReq.DeviceID, certsclient.OUIoTDevice,
		ownerPubKey, srv.caCert, srv.caKey,
		time.Now().Add(-10*time.Second), int(srv.deviceCertValidityDays))
	if err != nil {
		srv.tlsServer.WriteBadRequest(resp, fmt.Sprintf("ServeProvisionRequest: Failed creating client cert for %s: %s", provReq.DeviceID, err))
		return
	}
	// save the certificate using the device ID as the name
	// If the deviceID contains invalid characters this could fail
	newCertFile := provReq.DeviceID + "Cert.pem"
	if srv.certStore != "" {
		pemPath := path.Join(srv.certStore, newCertFile)
		err = certsclient.SaveX509CertToPEM(newClientCert, pemPath)
		if err != nil {
			srv.tlsServer.WriteInternalError(resp, fmt.Sprintf("ServeProvisionRequest: Failed creating client cert for %s: %s", provReq.DeviceID, err))
			return
		}
	}

	// return the certificate
	clientCertPEM := certsclient.X509CertToPEM(newClientCert)
	respStatus := idprovclient.PostProvisionResponseMessage{
		Status:        idprovclient.ProvisionStatusApproved,
		CaCertPEM:     string(srv.directory.CaCertPEM),
		ClientCertPEM: clientCertPEM,
	}
	// Create a signature. Don't use the secret if the certificate was valid.
	serialized, _ := json.Marshal(respStatus)
	// if cert is valid, peerCert points to the valid certificate
	if validCert {
		mySecret = ""
		logrus.Infof("ServeProvisionRequest: Authorized client '%s' (%s) for certificate generation/renewal of IoT device '%s'. Validity=%d days",
			peerCert.Subject.CommonName, peerCert.Subject.OrganizationalUnit, provReq.DeviceID, srv.deviceCertValidityDays)
	} else {
		logrus.Infof("ServeProvisionRequest: Authorized certificate generation for device '%s' using OOB verification. Validity is %d days",
			provReq.DeviceID, srv.deviceCertValidityDays)
	}
	signature, _ := idprovclient.Sign(string(serialized), mySecret)
	respStatus.Signature = signature
	responseMsg, _ := json.Marshal(respStatus)
	resp.Write(responseMsg)
	// remove the secret as it can only be used once
	srv.oobSecrets[provReq.DeviceID] = ""
}

// validateCertificate validates whether the given certificate allows provisioning.
// This assumes that the provided certificates are signed by the CA and that this check
// has already been performed by the TLS connection.
// * plugin and admin certificates are authorized to request certificate
// * IoT devices can renew their own certificate while it is still valid within the grace period
//  peerCert is the certificate to valid
//  deviceID is the device the certificate should be of when ou=IoTDevice
// This returns nil if the certificate is authorized to provision a new certificate or error when not
func (srv *IDProvServer) validateCertificate(peerCert *x509.Certificate, deviceID string) error {
	// Determine the OU with the highest permissions
	highestOU := certsclient.OUNone
	for _, ou := range peerCert.Subject.OrganizationalUnit {
		if ou == certsclient.OUAdmin || ou == certsclient.OUPlugin {
			highestOU = ou
			break
		} else if ou == certsclient.OUIoTDevice {
			highestOU = ou
		}
	}
	// No need to check for expiry as server rejects expired certificates

	// Admin and plugin certificates have no further requirements
	if highestOU == certsclient.OUAdmin || highestOU == certsclient.OUPlugin {
		// nothing to do
	} else if highestOU == certsclient.OUIoTDevice {
		// IoT device certificate renewal requested deviceID must match the certificate deviceID (CN)
		if peerCert.Subject.CommonName != deviceID {
			err := fmt.Errorf("validateCertificate: '%s' certificate of client %s is not for device %s",
				highestOU, peerCert.Subject.CommonName, deviceID)
			return err
		}
	} else {
		err := fmt.Errorf("validateCertificate: '%s' certificate of client %s is not authorized",
			highestOU, peerCert.Subject.CommonName)
		return err
	}

	return nil
}
