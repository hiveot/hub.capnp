package capnp4POGS

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/launcher"
)

// InfoListPOGS2Capnp converts a list of ServiceInfo type from POGS to Capnp
func InfoListPOGS2Capnp(infoList []launcher.ServiceInfo) hubapi.ServiceInfo_List {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	infoListCapnp, _ := hubapi.NewServiceInfo_List(seg, int32(len(infoList)))
	for i, serviceInfo := range infoList {
		serviceInfoCapnp := ServiceInfoPOGS2Capnp(serviceInfo)
		_ = infoListCapnp.Set(i, serviceInfoCapnp)
	}
	return infoListCapnp
}

// InfoListCapnp2POGS converts a list of ServiceInfo type from Capnp to POGS
func InfoListCapnp2POGS(infoListCapnp hubapi.ServiceInfo_List) []launcher.ServiceInfo {
	infoListPOGS := make([]launcher.ServiceInfo, infoListCapnp.Len())

	for i := 0; i < infoListCapnp.Len(); i++ {
		serviceInfoCapnp := infoListCapnp.At(i)
		serviceInfoPOGS := ServiceInfoCapnp2POGS(serviceInfoCapnp)
		infoListPOGS[i] = serviceInfoPOGS
	}

	return infoListPOGS
}
