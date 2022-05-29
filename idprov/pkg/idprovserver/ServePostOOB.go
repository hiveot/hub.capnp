package idprovserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/wost-go/pkg/certsclient"
)

// ServePostOOB receives out of band secret from the administrator
// The caller must have a valid client certificate with the OU admin or plugin
// The OOB secret is only valid for a 'short time' (days?) so it can be kept in memory
func (srv *IDProvServer) ServePostOOB(userID string, resp http.ResponseWriter, req *http.Request) {
	var oobMsg idprovclient.PostOobSecretMessage
	var deviceID string = "unknown"

	// Message body must be correct
	body, err := ioutil.ReadAll(req.Body)
	if err == nil {
		err = json.Unmarshal(body, &oobMsg)
	}
	if err != nil {
		srv.tlsServer.WriteBadRequest(resp, fmt.Sprintf("ServePostOOB: Invalid body in OOB message: %s", err))
		return
	}
	// Client must have a peer certificate to post an OOB secret
	deviceID = oobMsg.DeviceID
	hasCert := len(req.TLS.PeerCertificates)
	if hasCert == 0 {
		srv.tlsServer.WriteUnauthorized(resp, "ServePostOOB: Missing client certificate")
		return
	}
	peerCert := req.TLS.PeerCertificates[0]
	clientID := string(peerCert.Subject.CommonName)

	// Admin or plugins have permission
	hasPermission := false
	highestOU := certsclient.OUNone
	for _, ou := range peerCert.Subject.OrganizationalUnit {
		highestOU = ou
		if ou == certsclient.OUAdmin || ou == certsclient.OUPlugin {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		errMsg := fmt.Sprintf("ServePostOOB: Client %s is in OU %s with insufficient permissions", clientID, highestOU)
		srv.tlsServer.WriteUnauthorized(resp, errMsg)
		return
	}

	logrus.Infof("ServePostOOB: by client '%s' (ou=%s) for device '%s'", clientID, highestOU, deviceID)

	// The user has a client certificate with an admin serialnr. Approved
	srv.oobSecrets[oobMsg.DeviceID] = oobMsg.Secret
}
