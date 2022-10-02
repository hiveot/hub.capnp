// Package caphelp with helper for using capnp string TextList
package caphelp

import (
	"capnproto.org/go/capnp/v3"
)

// StringsToCapnp convert a string array to a capnp TextList
func StringsToCapnp(arr []string) capnp.TextList {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	tlist, _ := capnp.NewTextList(seg, int32(len(arr)))

	for i := 0; i < len(arr); i++ {
		tlist.Set(i, arr[i])
	}

	return tlist
}

// CapnpToStrings convert a capnp TextList to a string array
// errors are ignored
func CapnpToStrings(tlist capnp.TextList) []string {
	arr := make([]string, tlist.Len())
	for i := 0; i < tlist.Len(); i++ {
		text, _ := tlist.At(i)
		arr[i] = text
	}
	return arr
}
