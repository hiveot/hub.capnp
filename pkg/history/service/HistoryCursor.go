package service

import (
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub.go/pkg/vocab"
	"github.com/hiveot/hub/pkg/bucketstore"
)

// The HistoryCursor is a bucket cursor that converts the raw stored value to a ThingValue object
type HistoryCursor struct {
	thingAddr  string                    // the thing whose values this cursor is iterating
	filterName string                    // optional event name to filter on
	bc         bucketstore.IBucketCursor // the cursor of the underlying store
}

// convert the storage key and raw data to a thing value object
// this must match the encoding done in AddHistory
// This returns the value, or nil if the key is invalid
func (hc *HistoryCursor) decodeValue(key string, data []byte) (thingValue *thing.ThingValue, valid bool) {
	// key is constructed as  {timestamp}/{valueName}/{a|e}
	parts := strings.Split(key, "/")
	if len(parts) < 2 {
		return nil, false
	}
	millisec, _ := strconv.ParseInt(parts[0], 10, 64)
	ts := time.UnixMilli(millisec)
	timeIso8601 := ts.Format(vocab.ISO8601Format)
	thingValue = &thing.ThingValue{
		ThingAddr: hc.thingAddr,
		Name:      parts[1],
		ValueJSON: data,
		Created:   timeIso8601,
	}
	return thingValue, true
}

// findNextName iterates the cursor until the next value containing 'name' is found and the
// timestamp doesn't exceed untilTime.
// A successive call with an increased timestamp should return the next batch of results. Intended
// to iterated an hours/day/week at a time.
// This returns the next value, or nil if the value was not found.
//
//	name is the event name to match
//	until is the time not to exceed in the result. Intended to avoid unnecesary iteration in range queries
func (hc *HistoryCursor) findNextName(name string, until time.Time) (thingValue *thing.ThingValue, found bool) {
	found = false
	for {
		k, v, valid := hc.bc.Next()
		if !valid {
			// key is invalid. This means we reached the end of cursor
			return thingValue, false
		}
		// key is constructed as  {timestamp}/{valueName}/{a|e}
		parts := strings.Split(k, "/")
		if len(parts) != 3 {
			// key exists but is invalid. skip this entry
		} else {
			// check timestamp and name must match
			timestampmsec, _ := strconv.ParseInt(parts[0], 10, 64)
			if timestampmsec > until.UnixMilli() {
				// we passed the given time limit
				// undo the last step so that followup requests with a new time limit can include this result
				hc.bc.Prev()
				return thingValue, false
			}
			if name == parts[1] {
				// found a match. Decode and return it
				thingValue, found = hc.decodeValue(k, v)
				return
			}
			// name doesn't match. Skip this entry
		}
	}
}

// findPrevName iterates the cursor until the previous value containing 'name' is found and the
// timestamp is not before 'until' time.
// A successive call with an increased timestamp should return the next batch of results. Intended
// to iterated an hours/day/week at a time.
// This returns the previous value, or nil if the value was not found.
//
//	name is the event name to match
//	until is the time not to exceed in the result. Intended to avoid unnecesary iteration in range queries
func (hc *HistoryCursor) findPrevName(name string, until time.Time) (thingValue *thing.ThingValue, found bool) {
	found = false
	for {
		k, v, valid := hc.bc.Prev()
		if !valid {
			// key is invalid. This means we reached the beginning of cursor
			return thingValue, false
		}
		// key is constructed as  {timestamp}/{valueName}/{a|e}
		parts := strings.Split(k, "/")
		if len(parts) != 3 {
			// key exists but is invalid. skip this entry
		} else {
			// check timestamp and name must match
			timestampmsec, _ := strconv.ParseInt(parts[0], 10, 64)
			if timestampmsec < until.UnixMilli() {
				// we passed the given time limit
				// undo the last step so that followup requests with a new time limit can include this result
				hc.bc.Next()
				return thingValue, false
			}
			if name == parts[1] {
				// found a match. Decode and return it
				thingValue, found = hc.decodeValue(k, v)
				return
			}
			// name doesn't match. Skip this entry
		}
	}
}

// First returns the oldest value in the history
func (hc *HistoryCursor) First() (thingValue *thing.ThingValue, valid bool) {
	until := time.Now()
	k, v, valid := hc.bc.First()
	if !valid {
		// bucket is empty
		return thingValue, false
	}
	thingValue, valid = hc.decodeValue(k, v)
	if valid && hc.filterName != "" && thingValue.Name != hc.filterName {
		thingValue, valid = hc.findNextName(hc.filterName, until)
	}
	return thingValue, valid
}

// Last positions the cursor at the last key in the ordered list
func (hc *HistoryCursor) Last() (thingValue *thing.ThingValue, valid bool) {
	// the beginning of time?
	until := time.Time{}
	k, v, valid := hc.bc.Last()
	if !valid {
		// bucket is empty
		return thingValue, false
	}
	thingValue, valid = hc.decodeValue(k, v)
	if valid && hc.filterName != "" && thingValue.Name != hc.filterName {
		thingValue, valid = hc.findPrevName(hc.filterName, until)
	}
	return thingValue, valid
}

// Next moves the cursor to the next key from the current cursor
// First() or Seek must have been called first.
func (hc *HistoryCursor) Next() (thingValue *thing.ThingValue, valid bool) {
	until := time.Now()
	if hc.filterName != "" {
		thingValue, valid = hc.findNextName(hc.filterName, until)
		return thingValue, valid
	}
	k, v, valid := hc.bc.Next()
	if !valid {
		// at the end
		return thingValue, false
	}
	thingValue, valid = hc.decodeValue(k, v)
	return thingValue, valid
}

// NextN moves the cursor to the next N places from the current cursor
// and return a list with N values in incremental time order.
// itemsRemaining is false if the iterator has reached the end.
// Intended to speed up with batch iterations over rpc.
func (hc *HistoryCursor) NextN(steps uint) (values []*thing.ThingValue, itemsRemaining bool) {
	values = make([]*thing.ThingValue, 0, steps)
	// tbd is it faster to use NextN and sort the keys?
	for i := uint(0); i < steps; i++ {
		thingValue, valid := hc.Next()
		if !valid {
			break
		}
		values = append(values, thingValue)
	}
	return values, len(values) > 0
}

// Prev moves the cursor to the previous key from the current cursor
// Last() or Seek must have been called first.
func (hc *HistoryCursor) Prev() (thingValue *thing.ThingValue, valid bool) {
	until := time.Time{}
	if hc.filterName != "" {
		thingValue, valid = hc.findPrevName(hc.filterName, until)
		return thingValue, valid
	}

	k, v, valid := hc.bc.Prev()
	if !valid {
		// at begin
		return thingValue, false
	}

	thingValue, valid = hc.decodeValue(k, v)
	return thingValue, valid
}

// PrevN moves the cursor back N places from the current cursor
// and return a list with N values in decremental time order.
// itemsRemaining is true if the iterator has reached the beginning
// Intended to speed up with batch iterations over rpc.
func (hc *HistoryCursor) PrevN(steps uint) (values []*thing.ThingValue, itemsRemaining bool) {
	values = make([]*thing.ThingValue, 0, steps)
	// tbd is it faster to use NextN and sort the keys? - for a remote store yes
	for i := uint(0); i < steps; i++ {
		thingValue, valid := hc.Prev()
		if !valid {
			break
		}
		values = append(values, thingValue)
	}
	return values, len(values) > 0
}

// Release close the cursor and release its resources.
// This invalidates all values obtained from the cursor
func (hc *HistoryCursor) Release() {
	hc.bc.Release()
}

// Seek positions the cursor at the given searchKey and corresponding value.
// If the key is not found, the next key is returned.
// cursor.Close must be invoked after use in order to close any read transactions.
func (hc *HistoryCursor) Seek(isoTimestamp string) (thingValue *thing.ThingValue, valid bool) {
	until := time.Now()
	ts, err := dateparse.ParseAny(isoTimestamp)
	if err != nil {
		logrus.Infof("Seek using invalid timestamp '%s'. Thing address='%s'", isoTimestamp, hc.thingAddr)
		// can't do anything with this timestamp
		return
	}

	timeMilli := ts.UnixMilli()
	searchKey := strconv.FormatInt(timeMilli, 10) //+ "/" + thingValue.Name

	k, v, valid := hc.bc.Seek(searchKey)
	if !valid {
		// empty bucket
		return thingValue, false
	}
	thingValue, valid = hc.decodeValue(k, v)
	if valid && hc.filterName != "" && thingValue.Name != hc.filterName {
		thingValue, valid = hc.findNextName(hc.filterName, until)
	}

	return thingValue, valid
}

// NewHistoryCursor creates a new History Cursor for iterating the underlying bucket
//
//	thingAddr is the address the Thing can be reached at
//	filterName is an optional filter on value names, eg action, event, property or name
func NewHistoryCursor(thingAddr string, filterName string, bucketCursor bucketstore.IBucketCursor) *HistoryCursor {
	hc := &HistoryCursor{
		thingAddr:  thingAddr,
		bc:         bucketCursor,
		filterName: filterName,
	}
	return hc
}
