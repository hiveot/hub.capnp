package capnpserver

import (
	"context"

	"github.com/hiveot/hub/lib/thing"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/caphelp"
	"github.com/hiveot/hub/pkg/directory"
)

// ReadDirectoryCapnpServer provides the capnp RPC server for reading the directory
// This implements the capnproto generated interface ReadDirectory_Server
type ReadDirectoryCapnpServer struct {
	srv directory.IReadDirectory
}

func (capsrv *ReadDirectoryCapnpServer) Cursor(
	ctx context.Context, call hubapi.CapReadDirectory_cursor) error {

	cursor := capsrv.srv.Cursor(ctx)
	cursorSrv := NewDirectoryCursorCapnpServer(cursor)

	capability := hubapi.CapDirectoryCursor_ServerToClient(cursorSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCursor(capability)
	}
	return err
}
func (capsrv *ReadDirectoryCapnpServer) GetTD(ctx context.Context, call hubapi.CapReadDirectory_getTD) (err error) {
	var tv thing.ThingValue

	args := call.Args()
	publisherID, _ := args.PublisherID()
	thingID, _ := args.ThingID()
	tv, err = capsrv.srv.GetTD(ctx, publisherID, thingID)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		tvCapnp := caphelp.MarshalThingValue(tv)
		_ = res.SetTv(tvCapnp)
	}
	return err
}

//func (capsrv *ReadDirectoryCapnpServer) ListTDs(ctx context.Context, call hubapi.CapReadDirectory_listTDs) (err error) {
//	var tdList []string
//
//	args := call.Args()
//	limit := args.Limit()
//	offset := args.Offset()
//	tdList, err = capsrv.svc.ListTDs(ctx, int(limit), int(offset))
//	if err == nil {
//		res, _ := call.AllocResults()
//		textList, _ := res.NewTds(int32(len(tdList)))
//		for i := 0; i < len(tdList); i++ {
//			textList.Set(i, tdList[i])
//		}
//	}
//	return err
//}

// ListTDcb uses the capnp ability to pass callback interfaces. Great for sending streams of data.
//func (capsrv *ReadDirectoryCapnpServer) ListTDcb(ctx context.Context, call hubapi.CapReadDirectory_listTDcb) (err error) {
//	args := call.Args()
//	cb := args.Cb()
//
//	// the provided function implements the callback interface
//	err = capsrv.svc.ListTDcb(ctx, func(batch []string, isLast bool) error {
//		// send batches of TDs to the caller
//		// TODO: Do we need to create a new method for each batch? Can the callback be invoked repeatedly?
//		method, release := cb.Handler(ctx,
//			func(params hubapi.CapListCallback_handler_Params) error {
//				tdsCapnp := caphelp.MarshalStringList(batch)
//				err2 := params.SetTds(tdsCapnp)
//				params.SetIsLast(isLast)
//				return err2
//			})
//		defer release()
//		// the callback has no response
//		_, err3 := method.Struct()
//		return err3
//	})
//	return err
//}

//func (capsrv *ReadDirectoryCapnpServer) QueryTDs(ctx context.Context, call hubapi.CapReadDirectory_queryTDs) (err error) {
//	var jsonPath string
//	var tdList []string
//
//	args := call.Args()
//	limit := args.Limit()
//	offset := args.Offset()
//	jsonPath, err = args.JsonPath()
//	if err == nil {
//		tdList, err = capsrv.svc.QueryTDs(ctx, jsonPath, int(limit), int(offset))
//	}
//	if err == nil {
//		res, _ := call.AllocResults()
//		textList, _ := res.NewTds(int32(len(tdList)))
//		for i := 0; i < len(tdList); i++ {
//			textList.Set(i, tdList[i])
//		}
//	}
//	return err
//}

func (capsrv *ReadDirectoryCapnpServer) Shutdown() {
	// Release on the client calls capnp Shutdown.
	// Pass this to the server to cleanup
	capsrv.srv.Release()
}
