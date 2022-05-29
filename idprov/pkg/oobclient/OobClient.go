package oobclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/wostzone/wost-go/pkg/tlsclient"

	"github.com/wostzone/hub/idprov/pkg/idprovclient"

	"github.com/sirupsen/logrus"
)

// OOBClient admin tool to set the OOB secret
type OOBClient struct {
	addrPort   string // address and port of server
	client     *tlsclient.TLSClient
	clientCert *tls.Certificate
	caCert     *x509.Certificate
	running    bool
	directory  idprovclient.GetDirectoryMessage
}

// Directory provides the server directory
func (cl *OOBClient) Directory() idprovclient.GetDirectoryMessage {
	return cl.directory
}

// // Invoke a GET request from the server
// func (cl *OOBClient) GET(path string, payload interface{}) (result []byte, err error) {
// 	response, err := cl.client.Get(path, payload)
// 	return response, err
// }

// getDirectory obtains the IDProv server directory
// For use at startup
//  client is the http client to use
//  addr the server to connect to
func (cl *OOBClient) GetDirectory() (idprovclient.GetDirectoryMessage, error) {
	var dir idprovclient.GetDirectoryMessage

	data, err := cl.client.Get(idprovclient.IDProvDirectoryPath)
	if err != nil {
		logrus.Errorf("OOBClient.GetDirectory: Error %s", err)
		return dir, err
	}
	err = json.Unmarshal(data, &dir)
	if err != nil {
		logrus.Errorf("OOBClient.GetDirectory: Unmarshal Error %s", err)
		return dir, err
	}

	return dir, err
}

// Invoke a POST to the server. Intended for testing
func (cl *OOBClient) Post(path string, payload interface{}) (result []byte, err error) {
	response, err := cl.client.Post(path, payload)
	return response, err
}

// PostOOB posts the out of band provisioning secret
func (cl *OOBClient) PostOOB(deviceID string, secret string) (response []byte, err error) {
	oobMessage := idprovclient.PostOobSecretMessage{
		DeviceID: deviceID,
		Secret:   secret,
	}
	response, err = cl.client.Post(cl.directory.Endpoints.PostOobSecret, oobMessage)
	return response, err
}

// Start the OOB client and get the server directory.
func (cl *OOBClient) Start() (err error) {

	if cl.running {
		cl.Stop()
		logrus.Infof("OOBClient.Start: Restart with address %s", cl.addrPort)
	}
	// cl.addr = addr
	logrus.Infof("OOBClient.Start: Connecting to '%s' using certificate", cl.addrPort)

	cl.client = tlsclient.NewTLSClient(cl.addrPort, cl.caCert)

	// Should OOB client use an admin certificate instead?
	err = cl.client.ConnectWithClientCert(cl.clientCert)
	if err != nil {
		logrus.Error("OOBClient.Start: Failed to start: ", err)
		return err
	}

	// Get the post OOB URL from the server directory
	cl.directory, err = cl.GetDirectory()
	if err != nil {
		logrus.Error("OOBClient.Start: Failed to start: ", err)
		return err
	}

	cl.running = true
	return nil
}

//Stop the OOBClient client
func (cl *OOBClient) Stop() {
	if cl.running {
		cl.running = false
		cl.client.Close()
		cl.client = nil
	}
}

// NewOOBClient for provisioning out-of-band secret for IoT devices
//  addrPort is the server connection address with port
//  clientCert is the client TLS certificate to use to connect to the server
func NewOOBClient(addrPort string, clientCert *tls.Certificate, caCert *x509.Certificate) *OOBClient {
	c := OOBClient{
		addrPort:   addrPort,
		clientCert: clientCert,
		caCert:     caCert,
	}
	return &c
}
