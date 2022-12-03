package caphelp

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// TODO: generic Map conversion between POGS and capnp

// UnmarshalKeyValueMap deserializes a map of [key]value from a capnp message
// errors are ignored
func UnmarshalKeyValueMap(capMap hubapi.KeyValueMap) (valueMap map[string][]byte) {
	entries, _ := capMap.Entries()
	valueMap = make(map[string][]byte)

	for i := 0; i < entries.Len(); i++ {
		capEntry := entries.At(i)
		capKey, _ := capEntry.Key()
		capValue, _ := capEntry.Value()
		valueMap[capKey] = Clone(capValue)
	}
	return valueMap
}

// MarshalKeyValueMap serializes a key-value map to a capnp KeyValueMap
func MarshalKeyValueMap(valueMap map[string][]byte) hubapi.KeyValueMap {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capMap, _ := hubapi.NewKeyValueMap(seg)

	_, seg2, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capEntries, _ := hubapi.NewKeyValueMap_Entry_List(seg2, int32(len(valueMap)))
	i := 0
	for key, value := range valueMap {
		_, seg3, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		capEntry, _ := hubapi.NewKeyValueMap_Entry(seg3)
		_ = capEntry.SetKey(key)
		_ = capEntry.SetValue(value)
		_ = capEntries.Set(i, capEntry)
		i++
	}
	capMap.SetEntries(capEntries)

	return capMap
}
