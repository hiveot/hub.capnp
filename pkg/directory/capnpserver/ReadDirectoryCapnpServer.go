package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/directory"
)

// ReadDirectoryCapnpServer provides the capnp RPC server for reading the directory
// This implements the capnproto generated interface ReadDirectory_Server
type ReadDirectoryCapnpServer struct {
	srv directory.IReadDirectory
}

func (capsrv *ReadDirectoryCapnpServer) GetTD(ctx context.Context, call hubapi.CapReadDirectory_getTD) (err error) {
	var thingID string
	var td string

	args := call.Args()
	thingID, _ = args.ThingID()
	td, err = capsrv.srv.GetTD(ctx, thingID)
	if err == nil {
		res, _ := call.AllocResults()
		err = res.SetTdJson(td)
	}
	return err
}

func (capsrv *ReadDirectoryCapnpServer) ListTDs(ctx context.Context, call hubapi.CapReadDirectory_listTDs) (err error) {
	var tdList []string

	args := call.Args()
	limit := args.Limit()
	offset := args.Offset()
	tdList, err = capsrv.srv.ListTDs(ctx, int(limit), int(offset))
	if err == nil {
		res, _ := call.AllocResults()
		textList, _ := res.NewTds(int32(len(tdList)))
		for i := 0; i < len(tdList); i++ {
			textList.Set(i, tdList[i])
		}
	}
	return err
}

func (capsrv *ReadDirectoryCapnpServer) QueryTDs(ctx context.Context, call hubapi.CapReadDirectory_queryTDs) (err error) {
	var jsonPath string
	var tdList []string

	args := call.Args()
	limit := args.Limit()
	offset := args.Offset()
	jsonPath, err = args.JsonPath()
	if err == nil {
		tdList, err = capsrv.srv.QueryTDs(ctx, jsonPath, int(limit), int(offset))
	}
	if err == nil {
		res, _ := call.AllocResults()
		textList, _ := res.NewTds(int32(len(tdList)))
		for i := 0; i < len(tdList); i++ {
			textList.Set(i, tdList[i])
		}
	}
	return err
}
