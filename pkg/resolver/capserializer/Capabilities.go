package capserializer

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/resolver"
)

// MarshalCapabilityInfoList serializes CapabilityInfo list into a capnp list
func MarshalCapabilityInfoList(infoList []resolver.CapabilityInfo) (infoListCapnp hubapi.CapabilityInfo_List) {
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	infoListCapnp, _ = hubapi.NewCapabilityInfo_List(seg, int32(len(infoList)))
	for i, info := range infoList {
		infoCapnp := MarshalCapabilityInfo(&info)
		_ = infoListCapnp.Set(i, infoCapnp)
	}
	return infoListCapnp
}

// MarshalCapabilityInfo serializes CapabilityInfo into a capnp type
func MarshalCapabilityInfo(capInfo *resolver.CapabilityInfo) hubapi.CapabilityInfo {
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capInfoCapnp, _ := hubapi.NewCapabilityInfo(seg)
	if capInfo.CapabilityArgs != nil {
		capabilityArgsCapnp := caphelp.MarshalStringList(capInfo.CapabilityArgs)
		_ = capInfoCapnp.SetCapabilityArgs(capabilityArgsCapnp)
	}
	if capInfo.ClientTypes != nil {
		clientTypesCapnp := caphelp.MarshalStringList(capInfo.ClientTypes)
		_ = capInfoCapnp.SetClientTypes(clientTypesCapnp)
	}
	_ = capInfoCapnp.SetCapabilityName(capInfo.CapabilityName)
	_ = capInfoCapnp.SetProtocol(capInfo.Protocol)
	_ = capInfoCapnp.SetServiceID(capInfo.ServiceID)
	_ = capInfoCapnp.SetDNetwork(capInfo.DNetwork)
	_ = capInfoCapnp.SetDAddress(capInfo.DAddress)
	return capInfoCapnp
}

// UnmarshalCapabilyInfoList deserializes capnp CapabilityInfo list into a POGS type
func UnmarshalCapabilyInfoList(infoListCapnp hubapi.CapabilityInfo_List) (infoList []resolver.CapabilityInfo) {
	infoList = make([]resolver.CapabilityInfo, infoListCapnp.Len())
	var capInfoCapnp hubapi.CapabilityInfo
	for i := 0; i < infoListCapnp.Len(); i++ {
		capInfoCapnp = infoListCapnp.At(i)
		info := UnmarshalCapabilityInfo(capInfoCapnp)
		infoList[i] = info
	}
	return infoList
}

// UnmarshalCapabilityInfo deserializes capnp CapabilityInfo into a POGS type
func UnmarshalCapabilityInfo(infoCapnp hubapi.CapabilityInfo) (info resolver.CapabilityInfo) {

	clientTypesCapnp, _ := infoCapnp.ClientTypes()
	argsCapnp, _ := infoCapnp.CapabilityArgs()
	info.ClientTypes = caphelp.UnmarshalStringList(clientTypesCapnp)
	info.CapabilityArgs = caphelp.UnmarshalStringList(argsCapnp)
	info.CapabilityName, _ = infoCapnp.CapabilityName()
	info.Protocol, _ = infoCapnp.Protocol()
	info.ServiceID, _ = infoCapnp.ServiceID()
	info.DNetwork, _ = infoCapnp.DNetwork()
	info.DAddress, _ = infoCapnp.DAddress()
	return info
}
