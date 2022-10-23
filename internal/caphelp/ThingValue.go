// Package caphelp with helper convert between thing.ThingValue and capnp equivalent
package caphelp

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
)

// ThingValueCapnp2POGS converts a capnp thing value to a POGO thing.ThingValue
func ThingValueCapnp2POGS(capValue hubapi.ThingValue) (thingValue thing.ThingValue) {

	// errors are ignored. If these fails then there are bigger problems
	thingValue.ThingID, _ = capValue.ThingID()
	thingValue.Name, _ = capValue.Name()
	thingValue.ValueJSON, _ = capValue.ValueJSON()
	thingValue.Created, _ = capValue.Created()
	return thingValue
}

// ThingValueListCapnp2POGS convert a capnp ValueList to a POGO value array
// errors are ignored
func ThingValueListCapnp2POGS(tlist hubapi.ThingValue_List) []thing.ThingValue {
	arr := make([]thing.ThingValue, tlist.Len())
	for i := 0; i < tlist.Len(); i++ {
		capValue := tlist.At(i)
		arr[i] = ThingValueCapnp2POGS(capValue)
	}
	return arr
}

// ThingValueMapCapnp2POGS convert a capnp map to a POGO map with ThingValue objects
// errors are ignored
func ThingValueMapCapnp2POGS(capMap hubapi.ThingValueMap) (valueMap map[string]thing.ThingValue) {
	entries, _ := capMap.Entries()
	valueMap = make(map[string]thing.ThingValue)

	for i := 0; i < entries.Len(); i++ {
		capEntry := entries.At(i)
		capKey, _ := capEntry.Key()
		capValue, _ := capEntry.Value()
		thingValue := ThingValueCapnp2POGS(capValue)
		valueMap[capKey] = thingValue
	}
	return valueMap
}

// ThingValueListPOGS2Capnp convert an array from pog type to a capnp value type list
func ThingValueListPOGS2Capnp(arr []thing.ThingValue) hubapi.ThingValue_List {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capList, _ := hubapi.NewThingValue_List(seg, int32(len(arr)))

	for i := 0; i < len(arr); i++ {
		thingValue := arr[i]
		capValue := ThingValuePOGS2Capnp(thingValue)
		capList.Set(i, capValue)
	}

	return capList
}

// ThingValueMapPOGS2ToCapnp converts a map of thing.ThingValue to capnp equivalent
func ThingValueMapPOGS2ToCapnp(valueMap map[string]thing.ThingValue) hubapi.ThingValueMap {

	// errors are ignored. If these fails then there are bigger problems
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capMap, _ := hubapi.NewThingValueMap(seg)

	_, seg2, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capEntries, _ := hubapi.NewThingValueMap_Entry_List(seg2, int32(len(valueMap)))
	i := 0
	for name, thingValue := range valueMap {
		capValue := ThingValuePOGS2Capnp(thingValue)

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

// ThingValuePOGS2Capnp convert a POGO thing.ThingValue type to capnp defined thing value
// errors are ignored
func ThingValuePOGS2Capnp(thingValue thing.ThingValue) hubapi.ThingValue {

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
