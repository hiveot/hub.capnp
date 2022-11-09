// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// BucketCursorCapnpClient provides the POGS wrapper around the capnp client API
// This implements the IClientCursor interface
type BucketCursorCapnpClient struct {
	capability hubapi.CapBucketCursor // capnp cursor
}

// First positions the cursor at the first key in the ordered list
func (cl *BucketCursorCapnpClient) First() (key string, value []byte) {
	ctx := context.Background()
	method, release := cl.capability.First(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		key, _ = resp.Key()
		value, _ = resp.Value()
	}
	return key, value
}

// Last positions the cursor at the last key in the ordered list
func (cl *BucketCursorCapnpClient) Last() (key string, value []byte) {
	ctx := context.Background()
	method, release := cl.capability.Last(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		key, _ = resp.Key()
		value, _ = resp.Value()
	}
	return key, value
}

// Next moves the cursor to the next key from the current cursor
func (cl *BucketCursorCapnpClient) Next() (key string, value []byte) {
	ctx := context.Background()
	method, release := cl.capability.Next(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		key, _ = resp.Key()
		value, _ = resp.Value()
	}
	return key, value
}

// Prev moves the cursor to the previous key from the current cursor
func (cl *BucketCursorCapnpClient) Prev() (key string, value []byte) {
	ctx := context.Background()
	method, release := cl.capability.Prev(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		key, _ = resp.Key()
		value, _ = resp.Value()
	}
	return key, value
}

// Release the cursor capability
func (cl *BucketCursorCapnpClient) Release() {
	logrus.Infof("releasing bucket cursor")
	cl.capability.Release()
}

// Seek positions the cursor at the given searchKey and corresponding value.
// If the key is not found, the next key is returned.
// cursor.Close must be invoked after use in order to close any read transactions.
func (cl *BucketCursorCapnpClient) Seek(searchKey string) (key string, value []byte) {
	ctx := context.Background()
	method, release := cl.capability.Seek(ctx, func(params hubapi.CapBucketCursor_seek_Params) error {
		err2 := params.SetSearchKey(searchKey)
		return err2
	})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		key, _ = resp.Key()
		value, _ = resp.Value()
	}
	return key, value

}

// NewBucketCursorCapnpClient returns the capability to iterate a bucket
func NewBucketCursorCapnpClient(capability hubapi.CapBucketCursor) *BucketCursorCapnpClient {
	cl := &BucketCursorCapnpClient{
		capability: capability,
	}
	return cl
}
