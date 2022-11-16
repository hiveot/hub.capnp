package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/bucketstore"
)

// BucketCursorCapnpServer provides the capnp RPC server for a data cursor iterator.
// This implements the capnproto generated interface CapBucketCursor_Server
type BucketCursorCapnpServer struct {
	cursor bucketstore.IBucketCursor
}

func (srv BucketCursorCapnpServer) First(
	_ context.Context, call hubapi.CapBucketCursor_first) error {
	k, v, valid := srv.cursor.First()
	res, err := call.AllocResults()
	if err == nil {
		_ = res.SetKey(k)
		_ = res.SetValue(v)
		res.SetValid(valid)
	}
	return err
}

func (srv BucketCursorCapnpServer) Last(
	_ context.Context, call hubapi.CapBucketCursor_last) error {
	k, v, valid := srv.cursor.Last()
	res, err := call.AllocResults()
	_ = res.SetKey(k)
	_ = res.SetValue(v)
	res.SetValid(valid)
	return err
}

func (srv BucketCursorCapnpServer) Next(
	_ context.Context, call hubapi.CapBucketCursor_next) error {
	k, v, valid := srv.cursor.Next()
	res, err := call.AllocResults()
	_ = res.SetKey(k)
	_ = res.SetValue(v)
	res.SetValid(valid)
	return err
}

func (srv BucketCursorCapnpServer) NextN(
	_ context.Context, call hubapi.CapBucketCursor_nextN) error {
	args := call.Args()
	steps := args.Steps()
	docs, itemsRemaining := srv.cursor.NextN(uint(steps))
	res, err := call.AllocResults()
	if err != nil {
		docsCap := caphelp.MarshalKeyValueMap(docs)
		_ = res.SetDocs(docsCap)
		res.SetItemsRemaining(itemsRemaining)
	}
	return err
}

func (srv BucketCursorCapnpServer) Prev(
	_ context.Context, call hubapi.CapBucketCursor_prev) error {
	k, v, valid := srv.cursor.Prev()
	res, err := call.AllocResults()
	_ = res.SetKey(k)
	_ = res.SetValue(v)
	res.SetValid(valid)
	return err
}
func (srv BucketCursorCapnpServer) PrevN(
	_ context.Context, call hubapi.CapBucketCursor_prevN) error {
	args := call.Args()
	steps := args.Steps()
	docs, itemsRemaining := srv.cursor.PrevN(uint(steps))
	res, err := call.AllocResults()
	if err == nil {
		docsCap := caphelp.MarshalKeyValueMap(docs)
		_ = res.SetDocs(docsCap)
		res.SetItemsRemaining(itemsRemaining)
	}
	return err
}

func (srv BucketCursorCapnpServer) Seek(
	_ context.Context, call hubapi.CapBucketCursor_seek) error {
	args := call.Args()
	searchKey, _ := args.SearchKey()
	k, v, valid := srv.cursor.Seek(searchKey)
	res, err := call.AllocResults()
	_ = res.SetKey(k)
	_ = res.SetValue(v)
	res.SetValid(valid)
	return err
}

func (capsrv *BucketCursorCapnpServer) Shutdown() {
	// Release on the client calls capnp Shutdown.
	// Pass this to the server to cleanup
	capsrv.cursor.Release()
}

func NewBucketCursorCapnpServer(cursor bucketstore.IBucketCursor) *BucketCursorCapnpServer {
	cursorCapnpServer := &BucketCursorCapnpServer{
		cursor: cursor,
	}
	return cursorCapnpServer
}
