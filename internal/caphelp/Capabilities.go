package caphelp

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// CapabilityInfo provides information on a capabilities available through the gateway
type CapabilityInfo struct {

	// Name of the capability. This is the method name as defined by the service.
	CapabilityName string `json:"capabilityName"`

	// list of arguments that is required.
	CapabilityArgs []string

	// Type of clients for whom the capability is enabled. See ClientTypeXyz above
	ClientTypes []string `json:"clientType"`

	// Service name that is providing the capability.
	ServiceName string `json:"serviceName"`
}

// MarshalCapabilities serializes CapabilityInfo list into a capnp list
func MarshalCapabilities(infoList []CapabilityInfo) (infoListCapnp hubapi.CapabilityInfo_List) {
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	infoListCapnp, _ = hubapi.NewCapabilityInfo_List(seg, int32(len(infoList)))
	for i, info := range infoList {
		infoCapnp := MarshalCapabilityInfo(&info)
		_ = infoListCapnp.Set(i, infoCapnp)
	}
	return infoListCapnp
}

// MarshalCapabilityInfo serializes CapabilityInfo into a capnp type
func MarshalCapabilityInfo(capInfo *CapabilityInfo) hubapi.CapabilityInfo {
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	capInfoCapnp, _ := hubapi.NewCapabilityInfo(seg)
	if capInfo.CapabilityArgs != nil {
		capabilityArgsCapnp := MarshalStringList(capInfo.CapabilityArgs)
		_ = capInfoCapnp.SetCapabilityArgs(capabilityArgsCapnp)
	}
	if capInfo.ClientTypes != nil {
		clientTypesCapnp := MarshalStringList(capInfo.ClientTypes)
		_ = capInfoCapnp.SetClientTypes(clientTypesCapnp)
	}
	_ = capInfoCapnp.SetCapabilityName(capInfo.CapabilityName)
	_ = capInfoCapnp.SetServiceName(capInfo.ServiceName)
	return capInfoCapnp
}

// UnmarshalCapabilities deserializes capnp CapabilityInfo list into a POGS type
func UnmarshalCapabilities(infoListCapnp hubapi.CapabilityInfo_List) (infoList []CapabilityInfo) {
	infoList = make([]CapabilityInfo, infoListCapnp.Len())
	for i := 0; i < infoListCapnp.Len(); i++ {
		infoCapnp := infoListCapnp.At(i)
		info := UnmarshalCapabilityInfo(infoCapnp)
		infoList[i] = info
	}
	return infoList
}

// UnmarshalCapabilityInfo deserializes capnp CapabilityInfo into a POGS type
func UnmarshalCapabilityInfo(infoCapnp hubapi.CapabilityInfo) (info CapabilityInfo) {

	clientTypesCapnp, _ := infoCapnp.ClientTypes()
	argsCapnp, _ := infoCapnp.CapabilityArgs()
	info.ClientTypes = UnmarshalStringList(clientTypesCapnp)
	info.CapabilityArgs = UnmarshalStringList(argsCapnp)
	info.CapabilityName, _ = infoCapnp.CapabilityName()
	info.ServiceName, _ = infoCapnp.ServiceName()
	return info
}
