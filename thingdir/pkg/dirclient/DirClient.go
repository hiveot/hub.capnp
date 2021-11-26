// Package dirclient with client side functions to access the directory
package dirclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/td"
	"github.com/wostzone/hub/lib/client/pkg/tlsclient"
)

// Constants for use by server and applications
const (
	DefaultServiceName = "thingdir" // default discovery service name _{name}._tcp
	DefaultPort        = 8886       // default directory server listening port
)

// paths with REST commands
const (
	RouteThings  = "/things"           // list or query path
	RouteThingID = "/things/{thingID}" // for methods get, post, patch, delete
)

// query parameters
const (
	ParamOffset = "offset"
	ParamLimit  = "limit"
	ParamQuery  = "queryparams"
)

const DefaultLimit = 100
const MaxLimit = 1000

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

// ConnectWithLoginID open the connection to the directory server using a login ID and password for authentication
// For testing. Use the auth service instead
//  loginID  username or email
//  password credentials
func (dc *DirClient) ConnectWithLoginID(loginID string, password string) error {
	// TODO, support access token instead loginID
	accessToken, err := dc.tlsClient.ConnectWithLoginID(loginID, password)
	_ = accessToken
	return err
}

// Delete a TD.
func (dc *DirClient) Delete(id string) error {
	path := strings.Replace(RouteThingID, "{thingID}", id, 1)

	_, err := dc.tlsClient.Delete(path, nil)
	return err
}

// GetTD the TD with the given ID
//  id is the ThingID whose TD to get
func (dc *DirClient) GetTD(id string) (td td.ThingTD, err error) {

	path := strings.Replace(RouteThingID, "{thingID}", id, 1)
	resp, err := dc.tlsClient.Get(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &td)
	return td, err
}

// ListTDs
// Returns a list of TDs starting at the offset. The result is limited to the nr of records provided
// with the limit parameter. The server can choose to apply its own limit, in which case the lowest
// value is used.
//  offset of the list to query from
//  limit result to nr of TDs. Use 0 for default.
func (dc *DirClient) ListTDs(offset int, limit int) ([]td.ThingTD, error) {
	var tdList []td.ThingTD
	if limit == 0 {
		limit = DefaultLimit
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
func (dc *DirClient) PatchTD(id string, td td.ThingTD) error {
	var resp []byte
	var err error
	path := strings.Replace(RouteThingID, "{thingID}", id, 1)
	resp, err = dc.tlsClient.Patch(path, td)
	_ = resp
	return err
}

// QueryTDs with the given JSONPATH expression
// Returns a list of TDs matching the query, starting at the offset. The result is limited to the
// nr of records provided with the limit parameter. The server can choose to apply its own limit,
// in which case the lowest value is used.
//  offset of the list to query from
//  limit result to nr of TDs. Use 0 for default.
func (dc *DirClient) QueryTDs(jsonpath string, offset int, limit int) ([]td.ThingTD, error) {
	var tdList []td.ThingTD
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
func (dc *DirClient) UpdateTD(id string, td td.ThingTD) error {
	var resp []byte
	var err error
	path := strings.Replace(RouteThingID, "{thingID}", id, 1)
	resp, err = dc.tlsClient.Post(path, td)
	// resp, err = dc.tlsClient.Post(path, doc)
	_ = resp
	return err
}

// NewDirClient creates a new instance of the directory client
//  address is the listening address of the client
//  port to connect to
//  caCertPath server CA certificate for verification, obtained during provisioning using idprov
func NewDirClient(hostport string, caCert *x509.Certificate) *DirClient {
	tlsClient := tlsclient.NewTLSClient(hostport, caCert)
	dc := &DirClient{
		hostport:  hostport,
		tlsClient: tlsClient,
	}
	return dc
}
