// Package client that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/directory"
)

// ReadDirectoryCapnpClient is the POGO client to reading a directory
// It can only be obtained from the DirectoryCapnpClient
// This implements the IReadDirectory interface
type ReadDirectoryCapnpClient struct {
	capability hubapi.CapReadDirectory
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
		tds = caphelp.CapnpToStrings(capnpTDs)
	}
	return tds, err
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
		tds = caphelp.CapnpToStrings(capnpTDs)
	}
	return tds, err
}

func NewReadDirectoryCapnpClient(cap hubapi.CapReadDirectory) directory.IReadDirectory {
	return &ReadDirectoryCapnpClient{
		capability: cap,
	}
}
