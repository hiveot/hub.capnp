// Package caphelp with helper convert between thing.ThingValue and capnp equivalent
package caphelp

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
)

// CapnpToThingValue converts a capnp thing value to a POGO thing.ThingValue
func CapnpToThingValue(capValue hubapi.ThingValue) (thingValue thing.ThingValue) {

	thingValue.ThingID, _ = capValue.ThingID()
	thingValue.Name, _ = capValue.Name()
	thingValue.ValueJSON, _ = capValue.ValueJSON()
	thingValue.Created, _ = capValue.Created()
	return thingValue
}

// CapnpToThingValueList convert a capnp ValueList to a POGO value array
// errors are ignored
func CapnpToThingValueList(tlist hubapi.ThingValue_List) []thing.ThingValue {
	arr := make([]thing.ThingValue, tlist.Len())
	for i := 0; i < tlist.Len(); i++ {
		capValue := tlist.At(i)
		arr[i] = CapnpToThingValue(capValue)
	}
	return arr
}

// CapnpToThingValueMap convert a capnp map to a POGO map with ThingValue objects
// errors are ignored
func CapnpToThingValueMap(capMap hubapi.ThingValueMap) (valueMap map[string]thing.ThingValue) {
	entries, _ := capMap.Entries()
	valueMap = make(map[string]thing.ThingValue)

	for i := 0; i < entries.Len(); i++ {
		capEntry := entries.At(i)
		capKey, _ := capEntry.Key()
		capValue, _ := capEntry.Value()
		thingValue := CapnpToThingValue(capValue)
		valueMap[capKey] = thingValue
	}
	return valueMap
}

// ThingValueListToCapnp convert an array from pog type to a capnp value type list
func ThingValueListToCapnp(arr []thing.ThingValue) hubapi.ThingValue_List {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capList, _ := hubapi.NewThingValue_List(seg, int32(len(arr)))

	for i := 0; i < len(arr); i++ {
		thingValue := arr[i]
		capValue := ThingValueToCapnp(thingValue)
		capList.Set(i, capValue)
	}

	return capList
}

// ThingValueMapToCapnp converts a map of thing.ThingValue to capnp equivalent
func ThingValueMapToCapnp(valueMap map[string]thing.ThingValue) hubapi.ThingValueMap {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capMap, _ := hubapi.NewThingValueMap(seg)

	_, seg2, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capEntries, _ := hubapi.NewThingValueMap_Entry_List(seg2, int32(len(valueMap)))
	i := 0
	for name, thingValue := range valueMap {
		capValue := ThingValueToCapnp(thingValue)

		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		capEntry, _ := hubapi.NewThingValueMap_Entry(seg)
		capEntry.SetKey(name)
		capEntry.SetValue(capValue)
		capEntries.Set(i, capEntry)
		i++
	}
	capMap.SetEntries(capEntries)

	return capMap
}

// ThingValueToCapnp convert a POGO thing.ThingValue type to capnp defined thing value
// errors are ignored
func ThingValueToCapnp(thingValue thing.ThingValue) hubapi.ThingValue {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capValue, err := hubapi.NewThingValue(seg)
	if err == nil {
		capValue.SetThingID(thingValue.ThingID)
		capValue.SetName(thingValue.Name)
		capValue.SetValueJSON(thingValue.ValueJSON)
		capValue.SetCreated(thingValue.Created)
	}
	return capValue
}
