// Package dirclient with client side functions to access the directory
package dirclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wostzone/wost-go/pkg/thing"
	"github.com/wostzone/wost-go/pkg/tlsclient"

	"github.com/sirupsen/logrus"
)

// Constants for use by server and applications
const (
	DefaultServiceName = "thingdir" // default discovery service name _{name}._tcp
	DefaultPort        = 8886       // default directory server listening port
)

// paths REST routes for use in Gorilla Mux (https://github.com/gorilla/mux)
const (
	RouteThings       = "/things"           // list or query path
	RouteThingID      = "/things/{thingID}" // for methods get, post, patch, delete
	RouteThingsValues = "/values"           // things in query part
	RouteThingValues  = "/values/{thingID}" // values of thingID
)

// query parameters
//const (
//	ParamOffset       = "offset"
//	ParamLimit        = "limit"
//	ParamQuery        = "queryparams"
//	ParamUpdatedSince = "updatedSince"
//	ParamThings       = "things"
//)

// directory service query parameters
const (
	ParamPropNames = "propNames"
)

const DefaultLimit = 100
const MaxLimit = 1000

// PropValue hold property or event value including when it was last updated
type PropValue struct {
	// ISO8601 timestamp of when the value was last updated
	Updated string `json:"updated"`
	// The value
	Value interface{} `json:"value"`
}

// DirClient is a client for the WoST Directory service
// Intended for updating and reading TDs
type DirClient struct {
	hostport  string // address:port of the directory server, "" if unknown
	tlsClient *tlsclient.TLSClient
}

// Close the connection to the directory server
func (dc *DirClient) Close() {
	if dc.tlsClient != nil {
		dc.tlsClient.Close()
	}
}

// ConnectWithClientCert opens the connection to the directory server using a client certificate for authentication
//  clientCertFile  client certificate to authenticate the client with the broker
//  clientKeyFile   client key to authenticate the client with the broker
func (dc *DirClient) ConnectWithClientCert(tlsClientCert *tls.Certificate) error {
	err := dc.tlsClient.ConnectWithClientCert(tlsClientCert)
	return err
}

// ConnectWithJwtToken open the connection to the directory server using a login ID and given access token
//  loginID  username or email
//  accessToken JWT access token
func (dc *DirClient) ConnectWithJwtToken(loginID string, accessToken string) {
	dc.tlsClient.ConnectWithJwtAccessToken(loginID, accessToken)
}

// Delete a TD.
func (dc *DirClient) Delete(id string) error {
	path := strings.Replace(RouteThingID, "{thingID}", id, 1)

	_, err := dc.tlsClient.Delete(path, nil)
	return err
}

// GetThingValues returns the most recent values of properties and events of a thing
// Intended to get a recent snapshot of values before subscribing to value messages
// on the message bus.
//  id is the ThingID whose property to get
//  propName is the property to get
func (dc *DirClient) GetThingValues(thingID string) (
	values map[string]PropValue, err error) {

	path := strings.Replace(RouteThingValues, "{thingID}", thingID, 1)
	resp, err := dc.tlsClient.Get(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &values)
	return values, err
}

// GetTD the TD with the given ID
//  id is the ThingID whose TD to get
func (dc *DirClient) GetTD(id string) (td *thing.ThingTD, err error) {

	path := strings.Replace(RouteThingID, "{thingID}", id, 1)
	resp, err := dc.tlsClient.Get(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &td)
	return td, err
}

// GetThingsPropertyValues returns the property value of multiple things
//  thingIDs is a list of IDs whose properties to get
//  propNames is a list of the properties to get
// This returns a map of thingID's containing a map of property name-value pairs
func (dc *DirClient) GetThingsPropertyValues(thingIDs []string, propNames []string) (
	values map[string]map[string]PropValue, err error) {

	// specify things in query
	qThingIDs := fmt.Sprintf("?%s=%s", tlsclient.ParamThings, strings.Join(thingIDs, ","))
	qPropNames := fmt.Sprintf("&%s=%s", ParamPropNames, strings.Join(propNames, ","))
	resp, err := dc.tlsClient.Get(RouteThingsValues + qThingIDs + qPropNames)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &values)
	return values, err
}

// ListTDs
// Returns a list of TDs starting at the offset. The result is limited to the nr of records provided
// with the limit parameter. The server can choose to apply its own limit, in which case the lowest
// value is used.
//  offset of the list to query from
//  limit result to nr of TDs. Use 0 for default.
func (dc *DirClient) ListTDs(offset int, limit int) ([]thing.ThingTD, error) {
	var tdList []thing.ThingTD
	if limit == 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	path := fmt.Sprintf("%s?offset=%d&limit=%d", RouteThings, offset, limit)
	response, err := dc.tlsClient.Get(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(response, &tdList)
	logrus.Infof("DirClient.ListTDs. Returned %d TD(s)", len(tdList))
	return tdList, err
}

// PatchTD changes a TD with the attributes of the given TD
// Removed as this is not a client function
//func (dc *DirClient) PatchTD(id string, td thing.ThingTD) error {
//	var resp []byte
//	var err error
//	path := strings.Replace(RouteThingID, "{thingID}", id, 1)
//	resp, err = dc.tlsClient.Patch(path, td)
//	_ = resp
//	return err
//}

// QueryTDs with the given JSONPATH expression
// Returns a list of TDs matching the query, starting at the offset. The result is limited to the
// nr of records provided with the limit parameter. The server can choose to apply its own limit,
// in which case the lowest value is used.
//  offset of the list to query from
//  limit result to nr of TDs. Use 0 for default.
func (dc *DirClient) QueryTDs(jsonpath string, offset int, limit int) ([]thing.ThingTD, error) {
	var tdList []thing.ThingTD
	path := fmt.Sprintf("%s?queryparams=%s&offset=%d&limit=%d", RouteThings, jsonpath, offset, limit)
	response, err := dc.tlsClient.Get(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(response, &tdList)
	logrus.Infof("DirClient.QueryTDs. Returned %d TD(s)", len(tdList))
	return tdList, err
}

// UpdateTD updates the TD with the given ID, eg create/update
// Removed as this is not a client function
//func (dc *DirClient) UpdateTD(id string, td *thing.ThingTD) error {
//	var resp []byte
//	var err error
//	path := strings.Replace(RouteThingID, "{thingID}", id, 1)
//	resp, err = dc.tlsClient.Post(path, td)
//	// resp, err = dc.tlsClient.Post(path, doc)
//	_ = resp
//	return err
//}

// NewDirClient creates a new instance of the directory client
//  address is the listening address of the client
//  port to connect to
//  caCertPath server CA certificate for verification, obtained during provisioning using idprov
func NewDirClient(hostPort string, caCert *x509.Certificate) *DirClient {
	tlsClient := tlsclient.NewTLSClient(hostPort, caCert)
	dc := &DirClient{
		hostport:  hostPort,
		tlsClient: tlsClient,
	}
	return dc
}
