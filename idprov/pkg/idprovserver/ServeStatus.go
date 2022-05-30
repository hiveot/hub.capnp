package idprovserver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
)

// ServeStatus serves the provisioning status of a device
// This is pretty basic, the provisioning record is simply the existence of the
// certificate file for this device ID.
// The certificate filename is {deviceID}Cert.pem
func (srv *IDProvServer) ServeStatus(userID string, resp http.ResponseWriter, req *http.Request) {

	params := mux.Vars(req)
	deviceID := params["deviceID"]

	stat := idprovclient.GetDeviceStatusMessage{
		DeviceID:  deviceID,
		Status:    idprovclient.ProvisionStatusWaiting,
		CaCertPEM: string(srv.directory.CaCertPEM),
	}

	// Check if a certificate already exists
	clientCertFile := path.Join(srv.config.CertStoreFolder, deviceID+"Cert.pem")
	clientCertPEM, err := ioutil.ReadFile(clientCertFile)
	if err == nil {
		logrus.Infof("IdProvServer.ServeStatus. Certificate for device '%s' exists. Status is approved.", deviceID)
		// This client has an existing certificate, return it
		stat.Status = idprovclient.ProvisionStatusApproved
		stat.ClientCertPEM = clientCertPEM
		msg, _ := json.Marshal(stat)
		_, _ = resp.Write(msg)
		return
	}

	// If we have a secret but no certificate then we are waiting for the request or the secret to arrive
	_, found := srv.oobSecrets[deviceID]
	if found {
		logrus.Infof("IdProvServer.ServeStatus. OOB secret exists for device '%s'. Waiting for provisioning request.", deviceID)
		// waiting for a successful provisioning request
		stat.Status = idprovclient.ProvisionStatusWaiting
	} else {
		logrus.Infof("IdProvServer.ServeStatus. OOB secret does not exist for '%s'. Waiting for OOB secret.", deviceID)
		// wwaiting for an out of band secret to be provided
		stat.Status = idprovclient.ProvisionStatusWaiting
	}

	msg, _ := json.Marshal(stat)
	_, _ = resp.Write(msg)
}
