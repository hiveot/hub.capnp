package dirserver

import (
	"encoding/json"
	"fmt"
	"github.com/wostzone/hub/lib/client/pkg/certsclient"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/thingdir/pkg/dirclient"
)

// ServeThings lists or queries available TDs
// If a queryparam is provided then run a query, otherwise get the list
func (srv *DirectoryServer) ServeThings(userID string, response http.ResponseWriter, request *http.Request) {
	var offset = 0
	var tdList []interface{}
	certOU := certsclient.OUNone
	if len(request.TLS.PeerCertificates) > 0 {
		cert := request.TLS.PeerCertificates[0]
		if len(cert.Subject.OrganizationalUnit) > 0 {
			certOU = cert.Subject.OrganizationalUnit[0]
		}
	}

	limit, err := srv.tlsServer.GetQueryInt(request, dirclient.ParamLimit, dirclient.DefaultLimit)
	if limit > dirclient.MaxLimit {
		limit = dirclient.MaxLimit
	}
	if err == nil {
		offset, err = srv.tlsServer.GetQueryInt(request, dirclient.ParamOffset, 0)
	}
	if err != nil || offset < 0 {
		srv.tlsServer.WriteBadRequest(response, "ServeThings: offset or limit incorrect")
		return
	}
	jsonPath := srv.tlsServer.GetQueryString(request, dirclient.ParamQuery, "")

	aclFilter := NewAclFilter(userID, certOU, srv.authorizer)

	if jsonPath == "" {
		tdList = srv.store.List(offset, limit, aclFilter.FilterThing)
	} else {
		tdList, err = srv.store.Query(jsonPath, offset, limit, aclFilter.FilterThing)
		if err != nil {
			msg := fmt.Sprintf("ServeThings: query error: %s", err)
			srv.tlsServer.WriteBadRequest(response, msg)
			return
		}
	}

	logrus.Infof("ServeThings: user=%s (ou=%s), query='%s', offset=%d, limit=%d, #items=%d of %d",
		userID, certOU, jsonPath, offset, limit, len(tdList), srv.store.Size())
	msg, err := json.Marshal(tdList)
	if err != nil {
		msg := fmt.Sprintf("ServeThings: Marshal error %s", err)
		srv.tlsServer.WriteInternalError(response, msg)
		return
	}
	_, _ = response.Write(msg)
}
