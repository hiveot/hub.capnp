package dirserver

import (
	"encoding/json"
	"fmt"
	"github.com/wostzone/hub/authz/pkg/authorize"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// AclReadFilter determines read access to a thing TD. Intended for querying things.
// returns true if the userID has access to the thingID
// func FilterAclRead(string, thingID string) bool {
// 	return true
// }

// ServeTD serves a request for a TD document.
// This splits the request by its REST method: GET, POST, PUT, PATCH, DELETE, and authorizes
// before returning a result.
//
// * GET {thingID} returns the TD of a Thing
// * POST {thingID} replaces the TD of a Thing with the given TD
// * PUT {thingID} replaces the TD of a Thing with the given TD
// * PATCH {thingID} merges the TD of a Thing with the given TD
// * DELETE {thingID} removes the TD of a Thing
func (srv *DirectoryServer) ServeTD(userID string, response http.ResponseWriter, request *http.Request) {
	// determine the ID
	parts := strings.Split(request.URL.Path, "/")
	thingID := parts[len(parts)-1] // expect the thing ID
	certOU := srv.tlsServer.Authenticator().GetClientOU(request)

	logrus.Infof("ServeTD: %s for TD with ID %s", request.Method, thingID)
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
		!srv.authorizer(userID, certOU, thingID, authorize.AuthPubTD) {
		srv.tlsServer.WriteUnauthorized(response, "ServeGetTD: permission denied")
		return
	}

	tdMap, err := srv.dirStore.Get(thingID)
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
	if srv.authorizer != nil && !srv.authorizer(userID, certOU, thingID, authorize.AuthPubTD) {
		srv.tlsServer.WriteUnauthorized(response, "ServeDeleteTD: permission denied")
		return
	}

	srv.dirStore.Remove(thingID)
	// should we return the original? no, return 204
}

// ServePatchTD update only the provided parts of a thing's TD
func (srv *DirectoryServer) ServePatchTD(userID, certOU, thingID string, response http.ResponseWriter, request *http.Request) {

	if srv.authorizer != nil && !srv.authorizer(userID, certOU, thingID, authorize.AuthPubTD) {
		srv.tlsServer.WriteUnauthorized(response, "ServePatchTD: permission denied")
		return
	}

	tdMap := make(map[string]interface{})
	body, err := ioutil.ReadAll(request.Body)

	if err == nil {
		err = json.Unmarshal(body, &tdMap)
	}
	if err == nil {
		err = srv.dirStore.Patch(thingID, tdMap)
	}
	if err != nil {
		srv.tlsServer.WriteBadRequest(response, fmt.Sprintf("ServePatchTD: %s", err))
		return
	}
}

// ServeReplaceTD Creates or replace a TD
func (srv *DirectoryServer) ServeReplaceTD(userID, certOU, thingID string, response http.ResponseWriter, request *http.Request) {
	if srv.authorizer != nil && !srv.authorizer(userID, certOU, thingID, authorize.AuthPubTD) {
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
	existingTD, _ := srv.dirStore.Get(thingID)

	err = srv.dirStore.Replace(thingID, tdMap)
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
