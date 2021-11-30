// Package idprovclient with client to discover IDProv server and provision a certificate
package idprovclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/wostzone/hub/lib/client/pkg/certs"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/tlsclient"

	"github.com/sirupsen/logrus"
)

// DefaultPort is the default IDProv listening port
const DefaultPort = 8880

// IDProvClient IoT Device Provisioning client
type IDProvClient struct {
	clientID string // also used for client cert name

	clientCert *tls.Certificate  // client TLS certificate
	caCert     *x509.Certificate // CA certificate
	//
	caCertPath     string // path to store the received CA certificate
	clientCertPath string // path to store the received client certificate
	clientKeyPath  string // path to store the generated private key
	serverAddrPort string // idprov server address:port
	running        bool
	// tlsConfig  *tls.Config // config for new TLS connections
	// connection *tls.Conn
	// privKey      *ecdsa.PrivateKey
	publicKeyPEM string
	client       *tlsclient.TLSClient
	directory    GetDirectoryMessage
	// macAddr      string // outbound interface MAC address
}

// Directory provides the received server directory
func (cl *IDProvClient) Directory() GetDirectoryMessage {
	return cl.directory
}

// // ClientCert provides the client certificate, or nil if not available
// func (cl *IDProvClient) ClientCert() *tls.Certificate {
// 	return cl.client.Certificate()
// }

// PublicKeyPEM provides the client's public key
func (cl *IDProvClient) PublicKeyPEM() string {
	return cl.publicKeyPEM
}

// GetDirectory obtains the IDProv server directory
// Invoked at startup
func (cl *IDProvClient) GetDirectory() (GetDirectoryMessage, error) {
	var dir GetDirectoryMessage

	data, err := cl.client.Get(IDProvDirectoryPath)
	if err != nil {
		logrus.Errorf("IDProvClient.GetDirectory: Error %s", err)
		return dir, err
	}
	err = json.Unmarshal(data, &dir)
	if err != nil {
		logrus.Errorf("IDProvClient.GetDirectory: Unmarshal Error %s", err)
		return dir, err
	}

	return dir, err
}

// GetDeviceStatus obtains the provisioning status of a device
// GetDirectory must be called first.
func (cl *IDProvClient) GetDeviceStatus(deviceID string) (GetDeviceStatusMessage, error) {
	devStat := GetDeviceStatusMessage{}
	path := strings.ReplaceAll(cl.directory.Endpoints.GetDeviceStatus, "{deviceID}", deviceID)
	data, err := cl.client.Get(path)

	if err != nil {
		return devStat, err
	}
	// logrus.Infof("=----%s", data)
	err = json.Unmarshal(data, &devStat)
	if err != nil {
		logrus.Errorf("IDProvClient.GetDeviceStatus: Unmarshal Error %s", err)
		return devStat, err
	}

	return devStat, err
}

// Post invokes a POST request to the server. Intended for testing a bad payload
func (cl *IDProvClient) Post(path string, payload interface{}) (result []byte, err error) {
	response, err := cl.client.Post(path, payload)
	return response, err
}

// PostProvisioningRequest sends the provisioning request for a device.
// Start must be called first.
//
// This is intended for devices to obtain and renew a certificate
// plugin and admin clients to obtain a certificate on behalf of a device.
// The server approves this request only if the client has a matching OOB secret for the device,
// or the client uses an existing valid certificate for the device.
//
// Behavior:
//    1. Post the request
//    2. If approved,
//       2.1. verify the return signature against the provided signature. Only use the secret if we don't have a client cert
//       2.1. save the client certificate in the client certificate folder using the name {clientID}Cert.pem
//       2.2. restart the client so the certificate is used
//
//  deviceID of the device to request provisioning for. "" defaults to the client's device ID
//  secret that matches the OOB
func (cl *IDProvClient) PostProvisioningRequest(deviceID string, secret string) (*PostProvisionResponseMessage, error) {

	if cl.client == nil {
		err := fmt.Errorf("PostProvisioningRequest: Not Started (request) device %s", deviceID)
		logrus.Error(err)
		return nil, err
	}
	if deviceID == "" {
		deviceID = cl.clientID
	}
	provReqMessage := PostProvisionRequestMessage{
		DeviceID: deviceID,
		IP:       config.GetOutboundIP("").String(),
		// MAC:          myMAC,
		PublicKeyPEM: cl.publicKeyPEM,
		Signature:    "",
	}
	// Sign the message using the OOB secret
	serialized, _ := json.Marshal(provReqMessage)
	provReqMessage.Signature, _ = Sign(string(serialized), secret)

	response, err := cl.client.Post(cl.directory.Endpoints.PostProvisioningRequest, provReqMessage)
	if err != nil {
		logrus.Errorf("PostProvisioningRequest: Error %s", err)
		return nil, err
	}

	// get the response and verify the signature
	resp := PostProvisionResponseMessage{}
	err = json.Unmarshal(response, &resp)

	// response is unintelligable
	if err != nil {
		return nil, err
	}
	// Not yet approved, return info
	if resp.Status != ProvisionStatusApproved || resp.ClientCertPEM == "" || resp.CaCertPEM == "" {
		return &resp, nil
	}
	// Verify the signature to ensure the server has the same secret
	copyOfResp := resp
	copyOfResp.Signature = ""
	serialized, _ = json.Marshal(copyOfResp)
	if cl.client.Certificate() == nil {
		// verify with the secret if we're not using a client certificate
		err = Verify(string(serialized), secret, resp.Signature)
	} else {
		// If an existing certificate is refreshed, an OOB secret is not used.
		err = Verify(string(serialized), "", resp.Signature)
	}
	if err != nil {
		return nil, err
	}

	// save the CA and client certificates and restart the client
	logrus.Infof("PostProvisioningRequest: Approved, saving CA and Client certificates")
	err = ioutil.WriteFile(cl.caCertPath, []byte(resp.CaCertPEM), 0644)
	// Errors during saving the CA would have stopped start from succeeding
	if err == nil {
		err = ioutil.WriteFile(cl.clientCertPath, []byte(resp.ClientCertPEM), 0644)
	}
	if err != nil {
		err = fmt.Errorf("PostProvisioningRequest: Failed saving CA/client certificate: %s", err)
		logrus.Error(err)
		return &resp, err
	}

	// Reload the received certificate for use. This also checks if the cert is valid
	cl.caCert, err = certs.LoadX509CertFromPEM(cl.caCertPath)
	if err == nil {
		cl.clientCert, err = certs.LoadTLSCertFromPEM(cl.clientCertPath, cl.clientKeyPath)
	}
	if err != nil {
		err := fmt.Errorf("PostProvisioningRequest: Failed loading new CA/client certificate: %s", err)
		logrus.Error(err)
		return &resp, err
	}
	err = cl.restartClient(cl.clientCert, cl.caCert)
	return &resp, err
}

// Restart the TLS client in order to use an updated certificate
// This creates a new TLS client with the given client certificate
// The caller must persist the certificates in the clientCertPath and caCertPath
// Returns the new tls client
func (cl *IDProvClient) restartClient(clientCert *tls.Certificate, caCert *x509.Certificate) (err error) {
	if cl.client != nil {
		cl.client.Close()
	}
	cl.client = tlsclient.NewTLSClient(cl.serverAddrPort, caCert)
	if clientCert != nil {
		err = cl.client.ConnectWithClientCert(clientCert)
	} else {
		// no error until attempt to send
		cl.client.ConnectNoAuth()
	}
	if err != nil {
		logrus.Errorf("IDProvClient.RestartClient: Unable to start TLS client with address '%s': %s", cl.serverAddrPort, err)
		return err
	}
	return nil
}

// Start the IdProv client and read the server directory.
// If the client public/private keys do not yet exist they will be created for further use.
//
// Behavior:
//  1. If an ECDSA key does not exist in the certFolder then it will be created.
//  2. If a CA certificate exists in the certFolder then it will be used for server verification.
//  3. If a client certificate exist in the certFolder then the connection is configured for
// mutual authentication. This is required for renewing an existing certificate.
//  4. Invoke get directory on the server address. If it fails Start returns an error.
//  5. If no CA certificate exists the CA certificate included in the directory is saved.
//  6. If the CA was saved then restart the client so it is used.
func (cl *IDProvClient) Start() (err error) {
	var hasCA = false

	logrus.Infof("IDProvClient.Start: Connecting to address=%s", cl.serverAddrPort)
	if cl.running {
		cl.Stop()
	}
	if _, err := os.Stat(cl.caCertPath); err == nil {
		cl.caCert, err = certs.LoadX509CertFromPEM(cl.caCertPath)
		hasCA = err == nil
		if err != nil {
			logrus.Errorf("IDProvClient.Start. A CA cert was provided at %s but failed when loaded: %s", cl.caCertPath, err)
			// continue to recover and download a new cert.
			// return err
		}
	}

	// A private key will be created if it doesn't exist. It is needed for a public key to request a certificate
	// and for mutual auth TLS connection to renew a certificate.
	// pkPath := path.Join(certFolder, deviceKeyFile)
	privKey, err := certs.LoadKeysFromPEM(cl.clientKeyPath)
	if err != nil {
		privKey = certs.CreateECDSAKeys()
		err = certs.SaveKeysToPEM(privKey, cl.clientKeyPath)
		if err != nil {
			logrus.Errorf("IDProvClient.Start, failed saving private key: %s", err)
			return err
		}
		logrus.Infof("IDProvClient.Start, saved new private key to: %s", cl.clientKeyPath)
	}

	// If no server address was given, run discovery for up to 3 seconds to find one
	if cl.serverAddrPort == "" {
		cl.serverAddrPort, err = DiscoverProvisioningServer("", 3)
		if err != nil {
			logrus.Errorf("IDProvClient.Start: failed provisioning server discovery: %s", err)
			return err
		}
	}

	cl.publicKeyPEM, _ = certs.PublicKeyToPEM(&privKey.PublicKey)
	cl.clientCert, err = certs.LoadTLSCertFromPEM(cl.clientCertPath, cl.clientKeyPath)
	if err != nil {
		logrus.Infof("Loading client cert failed. Not using mutual auth.")
	}

	err = cl.restartClient(cl.clientCert, cl.caCert)
	if err != nil {
		// logrus.Errorf("IDProvClient.Start: Unable to start http client with address '%s:%d': %s", addr, port, err)
		return err
	}

	cl.directory, err = cl.GetDirectory()
	if err != nil {
		// logrus.Error("IDProvClient.Start: Unable to get directory: ", err)
		return err
	}

	// If no CA cert was available before, accept the one from the directory (leap of faith)
	// and restart so it is used.
	if !hasCA {
		// This assumes the provided directory contains a valid CA certificate. Should we test this?
		err = ioutil.WriteFile(cl.caCertPath, cl.directory.CaCertPEM, 0644)

		if err == nil {
			logrus.Infof("Restarting TLS client to use the new CA certificate")
			err = cl.restartClient(cl.clientCert, cl.caCert)
		}
	}

	cl.running = true
	return err
}

//Stop the IdProv client
func (cl *IDProvClient) Stop() {
	if cl.running {
		cl.running = false
		cl.client.Close()
		cl.client = nil
	}
}

// NewIDProvClient creates a client to obtain an authentication certificate from
// an IDProv provisioning server.
// The clientID must match the clientID provided to setup the out of band secret on the server.
// For provisioning to succeed both client ID and OOB secret must match. The clientID can
// be the MAC address or another ID available to the admin during OOB setup.
//
//  clientID the unique ID of this client.
//  serverAddrPort optional provisioning server address if known. Use "" to run discovery on Start
//  clientCertPath to store the received certificate in PEM format
//  clientKeyPath to store the generated client key in PEM format
//  caCertPath to store the received server CA in PEM format
func NewIDProvClient(clientID string, serverAddrPort string,
	clientCertPath, clientKeyPath string, caCertPath string) *IDProvClient {

	c := IDProvClient{
		serverAddrPort: serverAddrPort,
		clientID:       clientID,
		clientKeyPath:  clientKeyPath,
		clientCertPath: clientCertPath,
		caCertPath:     caCertPath,
	}
	return &c
}
