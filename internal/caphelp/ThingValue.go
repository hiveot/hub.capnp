// Package caphelp with helper convert between thing.ThingValue and capnp equivalent
package caphelp

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
)

// UnmarshalThingValue deserializes a ThingValue object from capnp
func UnmarshalThingValue(capValue hubapi.ThingValue) (thingValue thing.ThingValue) {

	// errors are ignored. If these fails then there are bigger problems
	thingValue.ThingID, _ = capValue.ThingID()
	thingValue.Name, _ = capValue.Name()
	thingValue.ValueJSON, _ = capValue.ValueJSON()
	thingValue.Created, _ = capValue.Created()
	return thingValue
}

// UnmarshalThingValueList deserializes a ThingValue array from capnp
// errors are ignored
func UnmarshalThingValueList(tlist hubapi.ThingValue_List) []thing.ThingValue {
	arr := make([]thing.ThingValue, tlist.Len())
	for i := 0; i < tlist.Len(); i++ {
		capValue := tlist.At(i)
		arr[i] = UnmarshalThingValue(capValue)
	}
	return arr
}

// UnmarshalThingValueMap deserializes a map of [key]ThingValue from a capnp message
// errors are ignored
func UnmarshalThingValueMap(capMap hubapi.ThingValueMap) (valueMap map[string]thing.ThingValue) {
	entries, _ := capMap.Entries()
	valueMap = make(map[string]thing.ThingValue)

	for i := 0; i < entries.Len(); i++ {
		capEntry := entries.At(i)
		capKey, _ := capEntry.Key()
		capValue, _ := capEntry.Value()
		thingValue := UnmarshalThingValue(capValue)
		valueMap[capKey] = thingValue
	}
	return valueMap
}

// MarshalThingValueList serializes a ThingValue array to a capnp list
func MarshalThingValueList(arr []thing.ThingValue) hubapi.ThingValue_List {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capList, _ := hubapi.NewThingValue_List(seg, int32(len(arr)))

	for i := 0; i < len(arr); i++ {
		thingValue := arr[i]
		capValue := MarshalThingValue(thingValue)
		capList.Set(i, capValue)
	}

	return capList
}

// MarshalThingValueMap serializes a map of thing.ThingValue to a capnp message
func MarshalThingValueMap(valueMap map[string]thing.ThingValue) hubapi.ThingValueMap {

	// errors are ignored. If these fails then there are bigger problems
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capMap, _ := hubapi.NewThingValueMap(seg)

	_, seg2, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capEntries, _ := hubapi.NewThingValueMap_Entry_List(seg2, int32(len(valueMap)))
	i := 0
	for name, thingValue := range valueMap {
		capValue := MarshalThingValue(thingValue)

		_, seg3, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		capEntry, _ := hubapi.NewThingValueMap_Entry(seg3)
		_ = capEntry.SetKey(name)
		_ = capEntry.SetValue(capValue)
		_ = capEntries.Set(i, capEntry)
		i++
	}
	_ = capMap.SetEntries(capEntries)

	return capMap
}

// MarshalThingValue serializes a thing.ThingValue object to a capnp message
// errors are ignored
func MarshalThingValue(thingValue thing.ThingValue) hubapi.ThingValue {

	// errors are ignored. If these fails then there are bigger problems
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capValue, err := hubapi.NewThingValue(seg)
	if err == nil {
		_ = capValue.SetThingID(thingValue.ThingID)
		_ = capValue.SetName(thingValue.Name)
		_ = capValue.SetValueJSON(thingValue.ValueJSON)
		_ = capValue.SetCreated(thingValue.Created)
	}
	return capValue
}
