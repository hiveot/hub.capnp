package dirserver

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/tlsclient"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"net/http"
	"strings"
	"time"
)

// ServeThingValues provides most recent property values of a Thing
// HTTP GET https://server:port/values/thingID?[&updatedSince=isodatetime][&propNames=propName1]
// Returns map with key-value pair of thing's property,
// Returns NotFound if thingID is not known or user is not authorized
func (srv *DirectoryServer) ServeThingValues(userID string, response http.ResponseWriter, request *http.Request) {
	resp := make(map[string]interface{})

	logrus.Infof("DirectoryServer.ServeThingValues: user=%s, URL=%s", userID, request.URL)
	certOU := srv.tlsServer.Authenticator().GetClientOU(request)

	thingID, found := mux.Vars(request)["thingID"]
	if !found {
		srv.tlsServer.WriteBadRequest(response, "ServeThingValues: missing thingID")
		return
	}
	// Any updated-since time?
	updatedSinceStr := srv.tlsServer.GetQueryString(request, tlsclient.ParamUpdatedSince, "")
	updatedSince, err := time.Parse(vocab.TimeFormat, updatedSinceStr)
	if updatedSinceStr == "" || err != nil {
		updatedSince = time.Unix(0, 0)
	}
	aclFilter := NewAclFilter(userID, certOU, srv.authorizer)
	resp = srv.GetPropValues(thingID, aclFilter, updatedSince)

	if resp == nil {
		srv.tlsServer.WriteNotFound(response, "ServeThingValues: Unknown thingID or not authorized")
	} else {
		valueResponse, _ := json.Marshal(resp)
		response.Write(valueResponse)
	}
}

// ServeMultipleThingsValues provides most recent property values
// HTTP GET https://server:port/values?things=thingID1,thingID2[&updatedSince=isodatetime][&propNames=propname1,...]
func (srv *DirectoryServer) ServeMultipleThingsValues(userID string, response http.ResponseWriter, request *http.Request) {
	//var offset = 0
	var thingsList []string
	resp := make(map[string]interface{})

	logrus.Infof("DirectoryServer.ServeMultipleThingsValues: user=%s, URL=%s", userID, request.URL)
	certOU := srv.tlsServer.Authenticator().GetClientOU(request)
	//
	//// offset and limit are optionally provided through query params
	//// this limits the nr of things in the response
	//limit, offset, err := srv.tlsServer.GetQueryLimitOffset(request, dirclient.DefaultLimit)
	//if err != nil || offset < 0 || limit < 0 {
	//	srv.tlsServer.WriteBadRequest(response, "ServeMultipleThingsValues: offset or limit incorrect")
	//	return
	//}

	// Any updated-since time?
	updatedSinceStr := srv.tlsServer.GetQueryString(request, tlsclient.ParamUpdatedSince, "")
	updatedSince, err := time.Parse(vocab.TimeFormat, updatedSinceStr)
	if updatedSinceStr == "" || err != nil {
		updatedSince = time.Unix(0, 0)
	}

	qThings := srv.tlsServer.GetQueryString(request, tlsclient.ParamThings, "")
	thingsList = strings.Split(qThings, ",")

	aclFilter := NewAclFilter(userID, certOU, srv.authorizer)
	// Get multiple things if specified in the query params
	for _, thingID := range thingsList {
		// only return props that are authorized
		props := srv.GetPropValues(thingID, aclFilter, updatedSince)
		if props != nil {
			resp[thingID] = props
		}
	}
	valueResponse, _ := json.Marshal(resp)
	response.Write(valueResponse)
}

// GetPropValues returns a map of property name-value pairs for the given Thing
// the aclFilter is used to authorize access to things properties
// returns nil if the thingID is not known or access is not authorized
func (srv *DirectoryServer) GetPropValues(thingID string, aclFilter AclFilter, updatedSince time.Time) map[string]interface{} {
	if !aclFilter.FilterThing(thingID) {
		return nil
	}
	values, found := srv.valueStore[thingID]
	if !found {
		// return nil if this is not a known thingID
		_, err := srv.dirStore.Get(thingID)
		if err != nil {
			return nil
		}
		// return result without properties
		return make(map[string]interface{})
	}
	return values
}
