// Package capnp4POGS with conversion of ProvisionStatus between canpnp and POGS
package capnp4POGS

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/provisioning"
)

// ProvStatusCapnp2POGS converts provisioning status from capnp struct to POGS
func ProvStatusCapnp2POGS(statusCapnp hubapi.ProvisionStatus) provisioning.ProvisionStatus {
	// errors are ignored. If these fails then there are bigger problems
	statusPOGS := provisioning.ProvisionStatus{}
	statusPOGS.DeviceID, _ = statusCapnp.DeviceID()
	statusPOGS.RequestTime, _ = statusCapnp.RequestTime()
	statusPOGS.RetrySec = int(statusCapnp.RetrySec())
	statusPOGS.Pending = statusCapnp.Pending()
	statusPOGS.ClientCertPEM, _ = statusCapnp.ClientCertPEM()
	statusPOGS.CaCertPEM, _ = statusCapnp.CaCertPEM()
	return statusPOGS
}

// ProvStatusPOGS2Capnp converts provisioning status from POGS to capnp struct
func ProvStatusPOGS2Capnp(statusPOGS provisioning.ProvisionStatus) hubapi.ProvisionStatus {
	// errors are ignored. If these fail then there are bigger problems
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	statusCapnp, _ := hubapi.NewProvisionStatus(seg)

	_ = statusCapnp.SetDeviceID(statusPOGS.DeviceID)
	_ = statusCapnp.SetRequestTime(statusPOGS.RequestTime)
	statusCapnp.SetRetrySec(int32(statusPOGS.RetrySec))
	statusCapnp.SetPending(statusPOGS.Pending)
	_ = statusCapnp.SetClientCertPEM(statusPOGS.ClientCertPEM)
	_ = statusCapnp.SetCaCertPEM(statusPOGS.CaCertPEM)
	return statusCapnp
}

// ProvStatusListCapnp2POGS converts provisioning status list from capnp struct to POGS
func ProvStatusListCapnp2POGS(statusListCapnp hubapi.ProvisionStatus_List) []provisioning.ProvisionStatus {
	// errors are ignored. If these fails then there are bigger problems
	statusListPOGS := make([]provisioning.ProvisionStatus, statusListCapnp.Len())
	for i := 0; i < statusListCapnp.Len(); i++ {
		statusCapnp := statusListCapnp.At(i)
		statusPOGS := ProvStatusCapnp2POGS(statusCapnp)
		statusListPOGS[i] = statusPOGS
	}
	return statusListPOGS
}

// ProvStatusListPOGS2Capnp converts a list of provisioning statuses from POGS to capnp struct
func ProvStatusListPOGS2Capnp(statusListPOGS []provisioning.ProvisionStatus) hubapi.ProvisionStatus_List {
	// errors are ignored. If these fail then there are bigger problems
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	statusListCapnp, _ := hubapi.NewProvisionStatus_List(seg, int32(len(statusListPOGS)))
	for i, statusPOGS := range statusListPOGS {
		statusCapnp := ProvStatusPOGS2Capnp(statusPOGS)
		statusListCapnp.Set(i, statusCapnp)
	}
	return statusListCapnp
}
