package capserializer

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/lib/caphelp"
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
	if capInfo.AuthTypes != nil {
		authTypesCapnp := caphelp.MarshalStringList(capInfo.AuthTypes)
		_ = capInfoCapnp.SetAuthTypes(authTypesCapnp)
	}
	capInfoCapnp.SetInterfaceID(capInfo.InterfaceID)
	capInfoCapnp.SetMethodID(capInfo.MethodID)
	_ = capInfoCapnp.SetInterfaceName(capInfo.InterfaceName)
	_ = capInfoCapnp.SetMethodName(capInfo.MethodName)
	_ = capInfoCapnp.SetProtocol(capInfo.Protocol)
	_ = capInfoCapnp.SetServiceID(capInfo.ServiceID)
	_ = capInfoCapnp.SetNetwork(capInfo.Network)
	_ = capInfoCapnp.SetAddress(capInfo.Address)
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

	authTypesCapnp, _ := infoCapnp.AuthTypes()
	info.AuthTypes = caphelp.UnmarshalStringList(authTypesCapnp)
	info.InterfaceID = infoCapnp.InterfaceID()
	info.MethodID = infoCapnp.MethodID()
	info.InterfaceName, _ = infoCapnp.InterfaceName()
	info.MethodName, _ = infoCapnp.MethodName()
	info.Protocol, _ = infoCapnp.Protocol()
	info.ServiceID, _ = infoCapnp.ServiceID()
	info.Network, _ = infoCapnp.Network()
	info.Address, _ = infoCapnp.Address()
	return info
}
