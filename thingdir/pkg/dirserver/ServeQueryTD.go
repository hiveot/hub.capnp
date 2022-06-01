package dirserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wostzone/wost-go/pkg/tlsclient"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/hub/thingdir/pkg/dirclient"
)

// ServeQueryTD queries available TDs
// If a queryparam is provided then run a query, otherwise get the list
func (srv *DirectoryServer) ServeQueryTD(userID string, response http.ResponseWriter, request *http.Request) {
	var offset = 0
	var tdList []interface{}
	certOU := srv.tlsServer.Authenticator().GetClientOU(request)

	// offset and limit are optionally provided through query params
	limit, offset, err := srv.tlsServer.GetQueryLimitOffset(request, dirclient.DefaultLimit)
	if err != nil || offset < 0 || limit < 0 {
		srv.tlsServer.WriteBadRequest(response, "ServeQueryTD: offset or limit incorrect")
		return
	}
	jsonPath := srv.tlsServer.GetQueryString(request, tlsclient.ParamQuery, "")

	aclFilter := NewAclFilter(userID, certOU, srv.authorizer)

	if jsonPath == "" {
		tdList = srv.dirStore.List(offset, limit, aclFilter.FilterThing)
	} else {
		tdList, err = srv.dirStore.Query(jsonPath, offset, limit, aclFilter.FilterThing)
		if err != nil {
			msg := fmt.Sprintf("ServeQueryTD: query error: %s", err)
			srv.tlsServer.WriteBadRequest(response, msg)
			return
		}
	}

	logrus.Infof("ServeQueryTD: user=%s (ou=%s), query='%s', offset=%d, limit=%d, #items=%d of %d",
		userID, certOU, jsonPath, offset, limit, len(tdList), srv.dirStore.Size())
	msg, err := json.Marshal(tdList)
	if err != nil {
		msg := fmt.Sprintf("ServeQueryTD: Marshal error %s", err)
		srv.tlsServer.WriteInternalError(response, msg)
		return
	}
	_, _ = response.Write(msg)
}
