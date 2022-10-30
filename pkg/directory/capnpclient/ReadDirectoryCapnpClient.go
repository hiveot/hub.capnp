// Package client that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/directory"
)

// ReadDirectoryCapnpClient is the POGS client to reading a directory
// It can only be obtained from the DirectoryCapnpClient
// This implements the IReadDirectory interface
type ReadDirectoryCapnpClient struct {
	capability hubapi.CapReadDirectory
}

func (cl *ReadDirectoryCapnpClient) Release() {
	cl.capability.Release()
}

// GetTD returns the TD document for the given Thing ID in JSON format
func (cl *ReadDirectoryCapnpClient) GetTD(ctx context.Context, thingID string) (tdJson string, err error) {

	method, release := cl.capability.GetTD(ctx,
		func(params hubapi.CapReadDirectory_getTD_Params) error {
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
func (cl *ReadDirectoryCapnpClient) ListTDs(ctx context.Context, limit int, offset int) (tds []string, err error) {
	method, release := cl.capability.ListTDs(ctx,
		func(params hubapi.CapReadDirectory_listTDs_Params) error {
			params.SetLimit(int32(limit))
			params.SetOffset(int32(offset))
			return nil
		})
	defer release()

	resp, err := method.Struct()
	if err == nil {
		capnpTDs, _ := resp.Tds()
		tds = caphelp.UnmarshalStringList(capnpTDs)
	}
	return tds, err
}

// ListTDReceiver implements the capnp 'server' for receiving callbacks.
// This is a capnp server for the client side.
// This implements the CapListCallback interface
type ListTDReceiver struct {
	pogsHandler func(batch []string, isLast bool) error
}

// Handler is a capnp callback method invoked by the server to push the list of TD's in batches
// This unmarshal's the arguments and pass it to the provided POGS handler.
func (receiver *ListTDReceiver) Handler(
	ctx context.Context, params hubapi.CapListCallback_handler) error {
	args := params.Args()
	tdsCapnp, _ := args.Tds()
	tds := caphelp.UnmarshalStringList(tdsCapnp)
	isLast := args.IsLast()
	_ = ctx
	receiver.pogsHandler(tds, isLast)
	return nil
}

// ListTDcb provides all TD documents in JSON format
func (cl *ReadDirectoryCapnpClient) ListTDcb(
	ctx context.Context, pogsHandler func(batch []string, isLast bool) error) (err error) {

	// The CapListCallback is a server that receives callbacks
	// This implements the CapListCallback API
	capHandler := &ListTDReceiver{pogsHandler: pogsHandler}
	// turn it into a capnp server
	capListCallback := hubapi.CapListCallback_ServerToClient(capHandler)

	// request the TD docs and provide a callback to pass the result
	method, release := cl.capability.ListTDcb(ctx,
		func(params hubapi.CapReadDirectory_listTDcb_Params) error {
			params.SetCb(capListCallback)
			return nil
		})
	defer release()

	_, err = method.Struct()
	return err
}

// QueryTDs returns the TD's filtered using JSONpath on the TD content
// See 'docs/query-tds.md' for examples
func (cl *ReadDirectoryCapnpClient) QueryTDs(ctx context.Context, jsonPath string, limit int, offset int) (tds []string, err error) {
	method, release := cl.capability.QueryTDs(ctx,
		func(params hubapi.CapReadDirectory_queryTDs_Params) error {
			params.SetJsonPath(jsonPath)
			params.SetLimit(int32(limit))
			params.SetOffset(int32(offset))
			return nil
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capnpTDs, _ := resp.Tds()
		tds = caphelp.UnmarshalStringList(capnpTDs)
	}
	return tds, err
}

func NewReadDirectoryCapnpClient(cap hubapi.CapReadDirectory) directory.IReadDirectory {
	return &ReadDirectoryCapnpClient{
		capability: cap,
	}
}
