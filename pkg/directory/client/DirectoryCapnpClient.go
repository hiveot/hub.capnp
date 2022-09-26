// Package client that wraps the capnp generated client with a POGS API
package client

import (
	"context"
	"net"
	"time"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
)

// DiriectoryCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IDirectoryStore interface
type DiriectoryCapnpClient struct {
	connection *rpc.Conn             // connection to capnp server
	capability hubapi.DirectoryStore // capnp client
	ctx        context.Context
	ctxCancel  context.CancelFunc
}

// GetTD returns the TD document for the given Thing ID in JSON format
func (cl *DiriectoryCapnpClient) GetTD(ctx context.Context, thingID string) (tdJson string, err error) {

	method, release := cl.capability.GetTD(cl.ctx,
		func(params hubapi.DirectoryStore_getTD_Params) error {
			err2 := params.SetThingID(thingID)
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		tdJson, err = resp.TdJson()
	}
	return tdJson, err
}

// ListTDs returns all TD documents in JSON format
func (cl *DiriectoryCapnpClient) ListTDs(ctx context.Context, limit int, offset int) (tds []string, err error) {
	method, release := cl.capability.ListTDs(cl.ctx,
		func(params hubapi.DirectoryStore_listTDs_Params) error {
			params.SetLimit(int32(limit))
			params.SetOffset(int32(offset))
			return nil
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capnpTDs, _ := resp.Tds()
		tds = caphelp.CapnpToStrings(capnpTDs)
	}
	return tds, err
}

// QueryTDs returns the TD's filtered using JSONpath on the TD content
// See 'docs/query-tds.md' for examples
func (cl *DiriectoryCapnpClient) QueryTDs(ctx context.Context, jsonPath string, limit int, offset int) (tds []string, err error) {
	method, release := cl.capability.QueryTDs(cl.ctx,
		func(params hubapi.DirectoryStore_queryTDs_Params) error {
			params.SetJsonPath(jsonPath)
			params.SetLimit(int32(limit))
			params.SetOffset(int32(offset))
			return nil
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capnpTDs, _ := resp.Tds()
		tds = caphelp.CapnpToStrings(capnpTDs)
	}
	return tds, err
}

// RemoveTD removes a TD document from the store
func (cl *DiriectoryCapnpClient) RemoveTD(ctx context.Context, thingID string) (err error) {
	method, release := cl.capability.RemoveTD(cl.ctx,
		func(params hubapi.DirectoryStore_removeTD_Params) error {
			params.SetThingID(thingID)
			return nil
		})
	defer release()
	_, err = method.Struct()
	return err
}

// UpdateTD updates the TD document in the directory
// If the TD with the given ID doesn't exist it will be added.
func (cl *DiriectoryCapnpClient) UpdateTD(ctx context.Context, thingID string, tdDoc string) (err error) {
	method, release := cl.capability.UpdateTD(cl.ctx,
		func(params hubapi.DirectoryStore_updateTD_Params) error {
			params.SetThingID(thingID)
			params.SetTdDoc(tdDoc)
			return nil
		})
	defer release()
	_, err = method.Struct()
	return err
}

// NewDirectoryStoreCapnpClient returns a directory store client using the capnp protocol
func NewDirectoryStoreCapnpClient(address string, isUDS bool) (*DiriectoryCapnpClient, error) {
	var cl *DiriectoryCapnpClient
	network := "tcp"
	if isUDS {
		network = "unix"
	}
	connection, err := net.Dial(network, address)
	if err == nil {
		transport := rpc.NewStreamTransport(connection)
		rpcConn := rpc.NewConn(transport, nil)
		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*60)
		capability := hubapi.DirectoryStore(rpcConn.Bootstrap(ctx))

		cl = &DiriectoryCapnpClient{
			connection: rpcConn,
			capability: capability,
			ctx:        ctx,
			ctxCancel:  ctxCancel,
		}
	}
	return cl, nil
}
