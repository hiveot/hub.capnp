package idprovserver_test

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/certs/pkg/certsetup"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/hub/idprov/pkg/oobclient"
	"github.com/wostzone/wost-go/pkg/certsclient"
)

//--- This uses TestMain in IDProvServer_test to start the server
//---
const testDeviceID = "device1"

// convenience function to create a signed client certificate for a device
func _createDeviceCert(deviceID string, ou string, startTime time.Time) (cert *x509.Certificate, key *ecdsa.PrivateKey, err error) {
	privKey := certsclient.CreateECDSAKeys()
	deviceCert, err := certsetup.CreateHubClientCert(
		testDeviceID, ou,
		&privKey.PublicKey, testCerts.CaCert, testCerts.CaKey, startTime, 1)
	// certs.SaveX509CertToPEM(deviceCert, device1CertPath)
	// certs.SaveKeysToPEM(privKey, device1KeyPath)

	return deviceCert, privKey, err
}

func TestProvisionByDeviceWithOOB(t *testing.T) {
	// A plugin issues the OOB secret
	oobClient := oobclient.NewOOBClient(idProvTestAddrPort, testCerts.PluginCert, testCerts.CaCert)
	err := oobClient.Start()
	require.NoError(t, err)
	_, err = oobClient.PostOOB(testDeviceID, clientOobSecret)
	assert.NoError(t, err)
	oobClient.Stop()

	// Device must provide matching OOB secret to provision
	removeDeviceCerts()
	idpc := idprovclient.NewIDProvClient(testDeviceID, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)
	err = idpc.Start()
	require.NoError(t, err)

	// expect status to be waiting
	//provStatus, err := idpc.GetDeviceStatus(testDeviceID)
	//assert.NoError(t, err)
	//assert.Equal(t, idprovclient.ProvisionStatusWaiting, provStatus.Status)

	// provisioning should now succeed
	provResp, err := idpc.PostProvisioningRequest("", clientOobSecret)
	require.NoError(t, err)
	assert.Equal(t, idprovclient.ProvisionStatusApproved, provResp.Status, "Expected status waiting for approval")
	assert.NotEmpty(t, provResp.ClientCertPEM)
	idpc.Stop()
}

func TestProvisionByDeviceFailWithoutOOB(t *testing.T) {
	testDeviceID2 := "device2nooob"
	// Start the client without a certificate. OOB must match instead
	removeDeviceCerts()
	idpc := idprovclient.NewIDProvClient(testDeviceID2, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)
	err := idpc.Start()
	require.NoError(t, err)

	// Response is waiting for OOB secret
	response, err := idpc.PostProvisioningRequest("", "")
	require.NoError(t, err)
	assert.Equal(t, idprovclient.ProvisionStatusWaiting, response.Status, "Expected waiting for oob")

	idpc.Stop()
}

func TestProvisionByDeviceWaitingForOOB(t *testing.T) {
	testDeviceID2 := "device2"
	// start the client without a certificate. OOB must match instead
	removeDeviceCerts()
	idpc := idprovclient.NewIDProvClient(testDeviceID, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)

	err := idpc.Start()
	require.NoError(t, err)
	// oob secret not yet submitted for client 2
	provStatus, err := idpc.PostProvisioningRequest(testDeviceID2, clientOobSecret)
	require.NoError(t, err)
	assert.Equal(t, idprovclient.ProvisionStatusWaiting, provStatus.Status, "Expected waiting for oob")

	idpc.Stop()
}
func TestProvisionByDeviceFailIncorrectOOBSecret(t *testing.T) {
	// plugin posts an OOB
	oobClient := oobclient.NewOOBClient(idProvTestAddrPort, testCerts.PluginCert, testCerts.CaCert)

	err := oobClient.Start()
	require.NoError(t, err)
	_, err = oobClient.PostOOB(testDeviceID, clientOobSecret)
	assert.NoError(t, err)
	oobClient.Stop()

	// device requests provisioning using OOB
	removeDeviceCerts()
	idpc := idprovclient.NewIDProvClient(testDeviceID, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)
	err = idpc.Start()
	require.NoError(t, err)

	provStatus, err := idpc.PostProvisioningRequest("", "incorrect secret")
	assert.NoError(t, err)
	assert.Equal(t, idprovclient.ProvisionStatusRejected, provStatus.Status, "Expected rejected with invalid secret")

	// retry with correct secret
	provStatus, err = idpc.PostProvisioningRequest("", clientOobSecret)
	assert.NoError(t, err)
	assert.Equal(t, idprovclient.ProvisionStatusApproved, provStatus.Status, "Expected approved for oob secret")

	idpc.Stop()
}

func TestProvisionByDeviceRenew(t *testing.T) {
	// clientCertFile := testDeviceID + "Cert.pem"
	removeDeviceCerts()

	// start with a device certificate so it can be renewed
	deviceCert, privKey, err := _createDeviceCert(testDeviceID, certsclient.OUIoTDevice, time.Now())
	require.NoError(t, err)
	certsclient.SaveX509CertToPEM(deviceCert, device1CertPath)
	certsclient.SaveKeysToPEM(privKey, device1KeyPath)

	// the client loads the created certificate
	idpc := idprovclient.NewIDProvClient(testDeviceID, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)
	err = idpc.Start()
	require.NoError(t, err)

	// // create a device client certificate signed by the CA for renewal
	// caCertPEM, _ := certsetup.LoadPEM(path.Join(serverCertFolder, certsetup.CaCertFile))
	// caKeyPEM, _ := certsetup.LoadPEM(path.Join(serverCertFolder, certsetup.CaKeyFile))
	// certPEM, err := certsetup.CreateClientCert(testDeviceID, certsetup.OUIoTDevice,
	// 	idpc.PublicKeyPEM(), caCertPEM, caKeyPEM, time.Now(), certsetup.TempCertDurationDays)
	// require.NoError(t, err)
	// certsetup.SaveCertToPEM(certPEM, path.Join(clientCertFolder, clientCertFile))
	// idpc.Stop()

	// refresh a valid client certificate. This does not need a secret as it has a valid client certificate.
	// idpc = idprovclient.NewIDProvClient(testDeviceID, idProvTestAddrPort,
	// 	device1CertPath, device1KeyPath, device1CaCertPath)
	// err = idpc.Start()
	// require.NoError(t, err)
	provResponse, err := idpc.PostProvisioningRequest("", "")
	require.NoError(t, err)
	assert.Equal(t, idprovclient.ProvisionStatusApproved, provResponse.Status)

	provStatus, err := idpc.GetDeviceStatus(testDeviceID)
	assert.NoError(t, err)
	assert.Equal(t, idprovclient.ProvisionStatusApproved, provStatus.Status)
	idpc.Stop()
}

func TestProvisionByDeviceRenewExpired(t *testing.T) {
	removeDeviceCerts()

	// start with a device certificate so it can be renewed
	deviceCert, privKey, err := _createDeviceCert(
		testDeviceID, certsclient.OUIoTDevice, time.Now().AddDate(0, 0, -3))
	require.NoError(t, err)
	certsclient.SaveX509CertToPEM(deviceCert, device1CertPath)
	certsclient.SaveKeysToPEM(privKey, device1KeyPath)
	certsclient.SaveX509CertToPEM(testCerts.CaCert, device1CaCertPath)

	// refresh client certificate as a Thing. No secret provided as we use
	// an existing client cert. As it is expired this will fail.
	// load the expired client certificate
	idpc := idprovclient.NewIDProvClient(testDeviceID, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)

	err = idpc.Start()
	require.Error(t, err)
	provStatus, err := idpc.PostProvisioningRequest("", "")
	_ = provStatus
	require.Error(t, err)
	idpc.Stop()
}

func TestProvisionByNoneOU(t *testing.T) {
	// clientCertFile := testDeviceID + "Cert.pem"

	removeDeviceCerts()

	// start with device certificate that is missing its OU
	deviceCert, privKey, err := _createDeviceCert(
		testDeviceID, certsclient.OUNone, time.Now())
	require.NoError(t, err)
	certsclient.SaveX509CertToPEM(deviceCert, device1CertPath)
	certsclient.SaveKeysToPEM(privKey, device1KeyPath)

	// start the client to generate a private key
	// idpc := idprovclient.NewIDProvClient(testDeviceID, idProvTestAddrPort,
	// 	device1CertPath, device1KeyPath, device1CaCertPath)
	// err := idpc.Start()
	// require.NoError(t, err)

	// // create a client certificate with unknown OU signed by the CA for renewal
	// caCertPEM, _ := ioutil.ReadFile(caCertPath)
	// caKeyPEM, _ := ioutil.ReadFile(caKeyPath)
	// certPEM, err := certsetup.CreateClientCert(testDeviceID, "",
	// 	idpc.PublicKeyPEM(), string(caCertPEM), string(caKeyPEM),
	// 	time.Now(), certsetup.TempCertDurationDays)
	// require.NoError(t, err)
	// certsetup.SaveCertToPEM(certPEM, path.Join(clientCertFolder, clientCertFile))

	// refresh client certificate as some client. Use client2 as it has no oob
	// restart the client to load the certificate
	idpc := idprovclient.NewIDProvClient(testDeviceID, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)
	err = idpc.Start()
	require.NoError(t, err)
	provStatus, err := idpc.PostProvisioningRequest("client2", clientOobSecret)
	// expect that the certificate is ignored and status is waiting (for oob)
	require.NoError(t, err)
	assert.Equal(t, idprovclient.ProvisionStatusWaiting, provStatus.Status)
	idpc.Stop()
}

// func TestProvisionCertificateExpiry(t *testing.T) {
// 	deviceID := "device1"
// 	removeDeviceCerts()
// 	pluginCertPath := path.Join(serverCertFolder, certsetup.PluginCertFile)
// 	pluginKeyPath := path.Join(serverCertFolder, certsetup.PluginKeyFile)
// 	idpClient := idprovclient.NewIDProvClient("plugin", idProvTestAddrPort,
// 		pluginCertPath, pluginKeyPath, caCertPath)
// 	idpClient.Start()

// 	provStatus, err := idpClient.PostProvisioningRequest(deviceID, "")
// 	require.NoError(t, err)
// 	assert.Equal(t, idprovclient.ProvisionStatusApproved, provStatus.Status, "Expected status waiting for approval")

// 	// test certificate validity
// 	clientCert, err := certs.LoadX509CertFromPEM(device1CertPath)
// 	assert.NoError(t, err)
// 	require.NotNil(t, clientCert, "Missing client certificate")
// 	validity := time.Until(clientCert.NotAfter)
// 	assert.Greater(t, validity, time.Hour*24, "Insufficient certificate validity period")
// 	idpClient.Stop()
// }

func TestProvisionInvalidBody(t *testing.T) {
	// plugin clients use the client cert in server cert folder
	idpc := idprovclient.NewIDProvClient("plugin", idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)
	err := idpc.Start()
	require.NoError(t, err)

	// Fail because message body is invalid
	_, err = idpc.Post(idpc.Directory().Endpoints.PostProvisioningRequest, "")
	require.Error(t, err)

	idpc.Stop()
}

func TestProvisionByPluginInvalidPubkey(t *testing.T) {
	// Device must provide matching OOB secret
	removeDeviceCerts()

	// A plugin issues the OOB secret
	oobClient := oobclient.NewOOBClient(idProvTestAddrPort, testCerts.PluginCert, testCerts.CaCert)
	err := oobClient.Start()
	require.NoError(t, err)
	_, err = oobClient.PostOOB(testDeviceID, clientOobSecret)
	assert.NoError(t, err)
	oobClient.Stop()

	// plugin clients with a bad publickey cert
	idpc := idprovclient.NewIDProvClient("plugin", idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)
	err = idpc.Start()
	require.NoError(t, err)

	provReqMessage := idprovclient.PostProvisionRequestMessage{
		DeviceID:     testDeviceID,
		IP:           "127.0.0.1",
		PublicKeyPEM: string("invalid pem"),
		Signature:    "",
	}
	serialized, _ := json.Marshal(provReqMessage)
	provReqMessage.Signature, _ = idprovclient.Sign(string(serialized), clientOobSecret)

	// Fail because request has a bad public key
	endpoint := idpc.Directory().Endpoints.PostProvisioningRequest
	_, err = idpc.Post(endpoint, provReqMessage)
	require.Error(t, err)
	idpc.Stop()
}

func TestProvisionByPluginInvalidDeviceID(t *testing.T) {
	// Device must provide matching OOB secret
	removeDeviceCerts()

	// start with a device certificate so it can be renewed
	deviceCert, privKey, err := _createDeviceCert(testDeviceID, certsclient.OUIoTDevice, time.Now())
	require.NoError(t, err)
	certsclient.SaveX509CertToPEM(deviceCert, device1CertPath)
	certsclient.SaveKeysToPEM(privKey, device1KeyPath)

	// plugin clients use the client cert in server cert folder
	idpc := idprovclient.NewIDProvClient(testDeviceID, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)
	err = idpc.Start()
	require.NoError(t, err)
	directory := idpc.Directory()

	// err = _createDeviceCert(testDeviceID, idpc.PublicKeyPEM(), time.Now())
	// assert.NoError(t, err)
	// restart to use the client cert
	// idpc.Stop()
	// idpc.Start()

	// Fail because deviceID is missing (amongst other reasons)
	provReqMessage := idprovclient.PostProvisionRequestMessage{
		// DeviceID: deviceID,
		Signature:    "fake",
		IP:           idProvTestAddr,
		PublicKeyPEM: idpc.PublicKeyPEM(),
	}
	_, err = idpc.Post(directory.Endpoints.PostProvisioningRequest, provReqMessage)
	assert.Error(t, err)

	// Again, this time with a different DeviceID. This client's cert is ignored
	// as it belongs to a different device. and a secret must be posted
	resp, err := idpc.PostProvisioningRequest("differentID", "")
	assert.Equal(t, idprovclient.ProvisionStatusWaiting, resp.Status)
	assert.NoError(t, err)

	stat, err := idpc.GetDeviceStatus("differentID")
	assert.Equal(t, idprovclient.ProvisionStatusWaiting, stat.Status)
	assert.NoError(t, err)

	idpc.Stop()
}

// func TestRenewByPluginExpiredCertificate(t *testing.T) {
// 	// generate a plugin authorized client certificate that has expired
// 	removeDeviceCerts()
// 	clientPrivKey, _ := certsetup.LoadOrCreateCertKey(clientCertFolder, certsetup.PluginKeyFile)
// 	clientPubKeyPEM, _ := signing.PublicKeyToPEM(&clientPrivKey.PublicKey)

// 	caCertPEM, _ := certsetup.LoadPEM(serverCertFolder, certsetup.CaCertFile)
// 	caKeyPEM, _ := certsetup.LoadPEM(serverCertFolder, certsetup.CaKeyFile)
// 	// Certificate is valid but expired
// 	certPEM, err := certsetup.CreateClientCert("", certsetup.OUClient,
// 		clientPubKeyPEM, caCertPEM, caKeyPEM, time.Now().AddDate(0, 0, -3), 1)
// 	require.NoError(t, err)
// 	certsetup.SaveCertToPEM(certPEM, clientCertFolder, testDeviceID+"Cert.pem")

// 	idpc := idprovclient.NewIDProvClient()
// 	// get directory should fail as cert has expired
// 	err = idpc.Start(testDeviceID, idProvTestAddr, idProvTestPort, clientCertFolder)
// 	assert.Error(t, err)

// 	// Expect certificate expired error - doesn't work as an expired cert fails to connect
// 	// provStatus, err := idpc.PostProvisioningRequest("", clientOobSecret)
// 	// _ = provStatus
// 	assert.Error(t, err)
// 	idpc.Stop()

// }
