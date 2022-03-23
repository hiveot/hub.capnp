package dirserver

import (
	"encoding/json"
	"fmt"
	"github.com/wostzone/hub/lib/client/pkg/certsclient"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/td"
)

// AclReadFilter determines read access to a thing TD. Intended for querying things.
// returns true if the userID has access to the thingID
// func FilterAclRead(string, thingID string) bool {
// 	return true
// }

// ServeThingByID serves a request for a particular Thing by its ID
// This splits the request by its REST method: GET, POST, PUT, PATCH, DELETE
func (srv *DirectoryServer) ServeThingByID(userID string, response http.ResponseWriter, request *http.Request) {
	// determine the ID
	parts := strings.Split(request.URL.Path, "/")
	thingID := parts[len(parts)-1] // expect the thing ID
	certOU := certsclient.OUNone
	if len(request.TLS.PeerCertificates) > 0 {
		cert := request.TLS.PeerCertificates[0]
		if len(cert.Subject.OrganizationalUnit) > 0 {
			certOU = cert.Subject.OrganizationalUnit[0]
		}
	}

	logrus.Infof("ServeThingByID: %s for TD with ID %s", request.Method, thingID)
	switch request.Method {
	case "GET":
		srv.ServeGetTD(userID, certOU, thingID, response)
	case "PATCH":
		srv.ServePatchTD(userID, certOU, thingID, response, request)
	case "POST":
		srv.ServeReplaceTD(userID, certOU, thingID, response, request)
	case "PUT":
		srv.ServeReplaceTD(userID, certOU, thingID, response, request)
	case "DELETE":
		srv.ServeDeleteTD(userID, certOU, thingID, response)
	default:
		srv.tlsServer.WriteBadRequest(response, fmt.Sprintf("Invalid method %s by %s", request.Method, userID))
	}
}

// ServeGetTD retrieve the requested TD
func (srv *DirectoryServer) ServeGetTD(userID, certOU, thingID string, response http.ResponseWriter) {

	if srv.authorizer != nil &&
		!srv.authorizer(userID, certOU, thingID, false, thing.MessageTypeTD) {
		srv.tlsServer.WriteUnauthorized(response, "ServeGetTD: permission denied")
		return
	}

	tdMap, err := srv.store.Get(thingID)
	if err != nil {
		msg := fmt.Sprintf("ServeGetTD: Unknown Thing with ID '%s'", thingID)
		srv.tlsServer.WriteNotFound(response, msg)
		return
	}
	msg, err := json.Marshal(tdMap)
	if err != nil {
		msg := fmt.Sprintf("ServeGetTD: Unable to marshal thing with ID %s", thingID)
		srv.tlsServer.WriteInternalError(response, msg)
		return
	}
	_, _ = response.Write(msg)
}

// ServeDeleteTD deletes the requested TD
func (srv *DirectoryServer) ServeDeleteTD(userID, certOU, thingID string, response http.ResponseWriter) {
	if srv.authorizer != nil && !srv.authorizer(userID, certOU, thingID, true, thing.MessageTypeTD) {
		srv.tlsServer.WriteUnauthorized(response, "ServeDeleteTD: permission denied")
		return
	}

	srv.store.Remove(thingID)
	// should we return the original? no, return 204
}

// ServePatchTD update only the provided parts of a thing's TD
func (srv *DirectoryServer) ServePatchTD(userID, certOU, thingID string, response http.ResponseWriter, request *http.Request) {

	if srv.authorizer != nil && !srv.authorizer(userID, certOU, thingID, true, thing.MessageTypeTD) {
		srv.tlsServer.WriteUnauthorized(response, "ServePatchTD: permission denied")
		return
	}

	tdMap := make(map[string]interface{})
	body, err := ioutil.ReadAll(request.Body)

	if err == nil {
		err = json.Unmarshal(body, &tdMap)
	}
	if err == nil {
		err = srv.store.Patch(thingID, tdMap)
	}
	if err != nil {
		srv.tlsServer.WriteBadRequest(response, fmt.Sprintf("ServePatchTD: %s", err))
		return
	}
}

// ServeReplaceTD Creates or replace a TD
func (srv *DirectoryServer) ServeReplaceTD(userID, certOU, thingID string, response http.ResponseWriter, request *http.Request) {
	if srv.authorizer != nil && !srv.authorizer(userID, certOU, thingID, true, thing.MessageTypeTD) {
		srv.tlsServer.WriteUnauthorized(response, "ServeReplaceTD: permission denied")
		return
	}

	tdMap := make(map[string]interface{})
	body, err := ioutil.ReadAll(request.Body)
	if err == nil {
		err = json.Unmarshal(body, &tdMap)
	}
	if err != nil {
		srv.tlsServer.WriteBadRequest(response, fmt.Sprintf("ServeReplaceTD: %s", err))
		return
	}
	existingTD, _ := srv.store.Get(thingID)

	err = srv.store.Replace(thingID, tdMap)
	if err != nil {
		srv.tlsServer.WriteBadRequest(response, fmt.Sprintf("ServeReplaceTD: %s", err))
		return
	}
	if existingTD != nil {
		// return 200 (OK)
		// default
	} else {
		// return 201 (Created)
		response.WriteHeader(http.StatusCreated)
	}
}
