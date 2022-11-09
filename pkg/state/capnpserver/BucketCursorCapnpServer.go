package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/state"
)

// BucketCursorCapnpServer provides the capnp RPC server for the store iterator.
// This implements the capnproto generated interface CapBucketCursor_Server
type BucketCursorCapnpServer struct {
	cursor state.IClientCursor
}

func (srv BucketCursorCapnpServer) First(
	_ context.Context, call hubapi.CapBucketCursor_first) error {
	k, v := srv.cursor.First()
	res, err := call.AllocResults()
	_ = res.SetKey(k)
	_ = res.SetValue(v)
	return err
}

func (srv BucketCursorCapnpServer) Last(
	_ context.Context, call hubapi.CapBucketCursor_last) error {
	k, v := srv.cursor.Last()
	res, err := call.AllocResults()
	_ = res.SetKey(k)
	_ = res.SetValue(v)
	return err
}

func (srv BucketCursorCapnpServer) Next(
	_ context.Context, call hubapi.CapBucketCursor_next) error {
	k, v := srv.cursor.Next()
	res, err := call.AllocResults()
	_ = res.SetKey(k)
	_ = res.SetValue(v)
	return err
}

func (srv BucketCursorCapnpServer) Prev(
	_ context.Context, call hubapi.CapBucketCursor_prev) error {
	k, v := srv.cursor.Prev()
	res, err := call.AllocResults()
	_ = res.SetKey(k)
	_ = res.SetValue(v)
	return err
}

func (srv BucketCursorCapnpServer) Seek(
	_ context.Context, call hubapi.CapBucketCursor_seek) error {
	args := call.Args()
	searchKey, _ := args.SearchKey()
	k, v := srv.cursor.Seek(searchKey)
	res, err := call.AllocResults()
	_ = res.SetKey(k)
	_ = res.SetValue(v)
	return err
}

func (capsrv *BucketCursorCapnpServer) Shutdown() {
	// Release on the client calls capnp Shutdown.
	// Pass this to the server to cleanup
	capsrv.cursor.Release()
}
