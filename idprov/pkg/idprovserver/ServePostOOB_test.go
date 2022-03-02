package idprovserver_test

import (
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/idprov/pkg/oobclient"
	"github.com/wostzone/hub/lib/client/pkg/certsclient"
)

//--- This uses TestMain in IDProvServer_test to start the server
//---

func TestPostOOB(t *testing.T) {
	const deviceID1 = "device1"
	const deviceSecret = "secret1"

	removeDeviceCerts()

	oobClient := oobclient.NewOOBClient(idProvTestAddrPort, testCerts.PluginCert, testCerts.CaCert)
	// OOB Client uses a plugin client certificate
	err := oobClient.Start()
	assert.NoError(t, err)

	// Posting an OOB secret means the device is approved (but not yet issued)
	resp, err := oobClient.PostOOB(deviceID1, deviceSecret)
	logrus.Infof("TestPostOOB: response='%s'", resp)
	assert.NoError(t, err)

	oobClient.Stop()

}

func TestServeOOBNoBody(t *testing.T) {
	// start the client without a certificate. OOB must match instead
	removeDeviceCerts()
	oobClient := oobclient.NewOOBClient(idProvTestAddrPort, testCerts.PluginCert, testCerts.CaCert)
	err := oobClient.Start()
	require.NoError(t, err)

	// use no message
	_, err = oobClient.Post(oobClient.Directory().Endpoints.PostOobSecret, "")
	require.Error(t, err)

	oobClient.Stop()
}

// Test the OOB client without using a certificate. This should fail
func TestServeOOBNoCert(t *testing.T) {
	const deviceID1 = "device1"
	removeDeviceCerts()
	oobClient := oobclient.NewOOBClient(idProvTestAddrPort, nil, testCerts.CaCert)
	err := oobClient.Start()
	require.Error(t, err)

	// use no message
	_, err = oobClient.PostOOB(deviceID1, "")
	require.Error(t, err)

	oobClient.Stop()
}

// Test the client with a certificate. OOB secret is not needed
func TestServeOOBClientCert(t *testing.T) {
	const clientID = "device1"

	// Fresh set of client keys
	removeDeviceCerts()
	// start with a device certificate so it can be renewed
	deviceCert, privKey, err := _createDeviceCert(testDeviceID, certsclient.OUIoTDevice, time.Now())
	require.NoError(t, err)
	certsclient.SaveX509CertToPEM(deviceCert, device1CertPath)
	certsclient.SaveKeysToPEM(privKey, device1KeyPath)

	// require.NoError(t, err)
	clientCertPEMPath := path.Join(clientCertFolder, clientID+"Cert.pem")
	clientKeyPEMPath := path.Join(clientCertFolder, clientID+"Key.pem")
	// certs.SaveTLSCertToPEM(newTLSCert, clientCertPEMPath, clientKeyPEMPath)

	clientCert, err := certsclient.LoadTLSCertFromPEM(clientCertPEMPath, clientKeyPEMPath)
	assert.NoError(t, err)
	// assert.NoError(t, err)
	// start client using the client cert and key
	oobClient := oobclient.NewOOBClient(idProvTestAddrPort, clientCert, testCerts.CaCert)
	err = oobClient.Start()
	require.NoError(t, err)

	// Fail. Devices cannot post their own oob. this has to be the admin or plugin
	_, err = oobClient.PostOOB(clientID, "secret")
	require.Error(t, err)

	oobClient.Stop()
}
