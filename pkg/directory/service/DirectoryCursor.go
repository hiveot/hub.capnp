package service

import (
	"encoding/json"
	"strings"

	"github.com/hiveot/hub/lib/thing"

	"github.com/hiveot/hub/pkg/bucketstore"
)

// The DirectoryCursor is a bucket cursor that converts the raw stored value to a ThingValue object
type DirectoryCursor struct {
	bc bucketstore.IBucketCursor // the cursor of the underlying store
}

// convert the storage key and raw data to a thing value object
// This returns the value, or nil if the key is invalid
func (dc *DirectoryCursor) decodeValue(key string, data []byte) (thingValue *thing.ThingValue, valid bool) {
	// key is constructed as  {timestamp}/{valueName}/{a|e}
	parts := strings.Split(key, "/")
	if len(parts) < 2 {
		return nil, false
	}
	thingValue = &thing.ThingValue{}
	_ = json.Unmarshal(data, thingValue)
	return thingValue, true
}

// First returns the oldest value in the history
func (dc *DirectoryCursor) First() (thingValue *thing.ThingValue, valid bool) {
	k, v, valid := dc.bc.First()
	if !valid {
		// bucket is empty
		return thingValue, false
	}
	thingValue, valid = dc.decodeValue(k, v)
	return thingValue, valid
}

// Next moves the cursor to the next key from the current cursor
// First() or Seek must have been called first.
func (dc *DirectoryCursor) Next() (thingValue *thing.ThingValue, valid bool) {
	k, v, valid := dc.bc.Next()
	if !valid {
		// at the end
		return thingValue, false
	}
	thingValue, valid = dc.decodeValue(k, v)
	return thingValue, valid
}

// NextN moves the cursor to the next N places from the current cursor
// and return a list with N values in incremental time order.
// itemsRemaining is false if the iterator has reached the end.
// Intended to speed up with batch iterations over rpc.
func (dc *DirectoryCursor) NextN(steps uint) (values []*thing.ThingValue, itemsRemaining bool) {
	values = make([]*thing.ThingValue, 0, steps)
	// tbd is it faster to use NextN and sort the keys?
	for i := uint(0); i < steps; i++ {
		thingValue, valid := dc.Next()
		if !valid {
			break
		}
		values = append(values, thingValue)
	}
	return values, len(values) > 0
}

// Release close the cursor and release its resources.
// This invalidates all values obtained from the cursor
func (dc *DirectoryCursor) Release() {
	dc.bc.Release()
}

// NewDirectoryCursor creates a new Cursor for iterating the directory entries
// gatewayID can be used to filter directory entries for the gateway only
func NewDirectoryCursor(bucketCursor bucketstore.IBucketCursor) *DirectoryCursor {
	dc := &DirectoryCursor{
		bc: bucketCursor,
	}
	return dc
}
