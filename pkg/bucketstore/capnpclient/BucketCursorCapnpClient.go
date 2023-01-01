// Package capnpclient that for using the bucket store cursor over capnp
package capnpclient

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/caphelp"
)

// BucketCursorCapnpClient provides a capnp RPC client of the bucket cursor
// This implements the IBucketCursor interface
type BucketCursorCapnpClient struct {
	capability hubapi.CapBucketCursor // capnp cursor
}

// First positions the cursor at the first key in the ordered list
func (cl *BucketCursorCapnpClient) First() (key string, value []byte, valid bool) {
	ctx := context.Background()
	method, release := cl.capability.First(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		key, _ = resp.Key()
		value, _ = resp.Value()
		valid = resp.Valid()
		// clone value as the capnp buffer is reused
		value = caphelp.Clone(value)
	}
	return
}

// Last positions the cursor at the last key in the ordered list
func (cl *BucketCursorCapnpClient) Last() (key string, value []byte, valid bool) {
	ctx := context.Background()
	method, release := cl.capability.Last(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		key, _ = resp.Key()
		value, _ = resp.Value()
		valid = resp.Valid()
		// clone value as the capnp buffer is reused
		value = caphelp.Clone(value)
	}
	return
}

// Next moves the cursor to the next key from the current cursor
func (cl *BucketCursorCapnpClient) Next() (key string, value []byte, valid bool) {
	ctx := context.Background()
	method, release := cl.capability.Next(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		key, _ = resp.Key()
		value, _ = resp.Value()
		valid = resp.Valid()
		// clone value as the capnp buffer is reused
		value = caphelp.Clone(value)
	}
	return
}

// NextN moves the cursor to the next N steps from the current cursor
func (cl *BucketCursorCapnpClient) NextN(steps uint) (docs map[string][]byte, itemsRemaining bool) {
	ctx := context.Background()

	method, release := cl.capability.NextN(ctx,
		func(params hubapi.CapBucketCursor_nextN_Params) error {
			params.SetSteps(uint32(steps))
			return nil
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capMap, _ := resp.Docs()
		docs = caphelp.UnmarshalKeyValueMap(capMap)
		itemsRemaining = resp.ItemsRemaining()
	}
	return
}

// Prev moves the cursor to the previous key from the current cursor
func (cl *BucketCursorCapnpClient) Prev() (key string, value []byte, valid bool) {
	ctx := context.Background()
	method, release := cl.capability.Prev(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		key, _ = resp.Key()
		value, _ = resp.Value()
		valid = resp.Valid()
		// clone value as the capnp buffer is reused
		value = caphelp.Clone(value)
	}
	return
}

// PrevN moves the cursor back N steps from the current cursor
func (cl *BucketCursorCapnpClient) PrevN(steps uint) (docs map[string][]byte, itemsRemaining bool) {
	ctx := context.Background()
	method, release := cl.capability.PrevN(ctx,
		func(params hubapi.CapBucketCursor_prevN_Params) error {
			params.SetSteps(uint32(steps))
			return nil
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capMap, _ := resp.Docs()
		docs = caphelp.UnmarshalKeyValueMap(capMap)
		itemsRemaining = resp.ItemsRemaining()
	}
	return
}

// Release the cursor capability
func (cl *BucketCursorCapnpClient) Release() {
	logrus.Infof("releasing bucket cursor")
	cl.capability.Release()
}

// Seek positions the cursor at the given searchKey and corresponding value.
// If the key is not found, the next key is returned.
// cursor.Close must be invoked after use in order to close any read transactions.
func (cl *BucketCursorCapnpClient) Seek(searchKey string) (key string, value []byte, valid bool) {
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
		valid = resp.Valid()
		// clone value as the capnp buffer is reused
		value = caphelp.Clone(value)
	}
	return

}

// NewBucketCursorCapnpClient returns the capability to iterate a bucket
func NewBucketCursorCapnpClient(capability hubapi.CapBucketCursor) *BucketCursorCapnpClient {
	cl := &BucketCursorCapnpClient{
		capability: capability,
	}
	return cl
}
