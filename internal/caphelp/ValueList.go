// Package caphelp with helper for using capnp value list
package caphelp

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
)

// ToCapnpValueList convert an array from pog type to a capnp value type list
func ToCapnpValueList(arr []thing.ThingValue) hubapi.ThingValue_List {

	msg, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	_ = msg
	//st, err := capnp.NewRootStruct(seg, capnp.ObjectSize{DataSize: 0, PointerCount: 2})

	capList, _ := hubapi.NewThingValue_List(seg, int32(len(arr)))
	for i := 0; i < len(arr); i++ {
		histValue := arr[i]
		capValue := capList.At(i)
		capValue.SetName(histValue.Name)
		capValue.SetValueJSON(histValue.ValueJSON)
		capValue.SetCreated(histValue.Created)
	}

	return capList
}
