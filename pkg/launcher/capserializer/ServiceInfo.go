package capserializer

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/launcher"
)

// MarshalServiceInfo serializes a ServiceInfo object to a capnp message
func MarshalServiceInfo(serviceInfoPOGS launcher.ServiceInfo) hubapi.ServiceInfo {
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	siCapnp, _ := hubapi.NewServiceInfo(seg)
	siCapnp.SetCpu(int32(serviceInfoPOGS.CPU))
	siCapnp.SetRss(int64(serviceInfoPOGS.RSS))
	_ = siCapnp.SetStatus(serviceInfoPOGS.Status)
	_ = siCapnp.SetModifiedTime(serviceInfoPOGS.ModifiedTime)
	_ = siCapnp.SetName(serviceInfoPOGS.Name)
	_ = siCapnp.SetPath(serviceInfoPOGS.Path)
	siCapnp.SetPid(int32(serviceInfoPOGS.PID))
	siCapnp.SetStartCount(int32(serviceInfoPOGS.StartCount))
	_ = siCapnp.SetStartTime(serviceInfoPOGS.StartTime)
	_ = siCapnp.SetStopTime(serviceInfoPOGS.StopTime)
	siCapnp.SetRunning(serviceInfoPOGS.Running)
	siCapnp.SetSize(serviceInfoPOGS.Size)
	siCapnp.SetUptime(int32(serviceInfoPOGS.Uptime))

	return siCapnp
}

// MarshalServiceInfoList serializes a list of ServiceInfo to a Capnp message
func MarshalServiceInfoList(infoList []launcher.ServiceInfo) hubapi.ServiceInfo_List {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	infoListCapnp, _ := hubapi.NewServiceInfo_List(seg, int32(len(infoList)))
	for i, serviceInfo := range infoList {
		serviceInfoCapnp := MarshalServiceInfo(serviceInfo)
		_ = infoListCapnp.Set(i, serviceInfoCapnp)
	}
	return infoListCapnp
}

// UnmarshalServiceInfo deserializes a ServiceInfo object from a capnp message
func UnmarshalServiceInfo(serviceInfoCapnp hubapi.ServiceInfo) launcher.ServiceInfo {
	status, _ := serviceInfoCapnp.Status()
	modifiedTime, _ := serviceInfoCapnp.ModifiedTime()
	serviceName, _ := serviceInfoCapnp.Name()
	servicePath, _ := serviceInfoCapnp.Path()
	startTime, _ := serviceInfoCapnp.StartTime()
	stopTime, _ := serviceInfoCapnp.StopTime()

	serviceInfoPOGS := launcher.ServiceInfo{
		CPU:          int(serviceInfoCapnp.Cpu()),
		RSS:          int(serviceInfoCapnp.Rss()),
		Status:       status,
		ModifiedTime: modifiedTime,
		Name:         serviceName,
		Path:         servicePath,
		PID:          int(serviceInfoCapnp.Pid()),
		StartCount:   int(serviceInfoCapnp.StartCount()),
		StartTime:    startTime,
		StopTime:     stopTime,
		Running:      serviceInfoCapnp.Running(),
		Size:         serviceInfoCapnp.Size(),
		Uptime:       int(serviceInfoCapnp.Uptime()),
	}
	return serviceInfoPOGS
}

// UnmarshalServiceInfoList deserializes a list of ServiceInfo from a Capnp message
func UnmarshalServiceInfoList(infoListCapnp hubapi.ServiceInfo_List) []launcher.ServiceInfo {
	infoListPOGS := make([]launcher.ServiceInfo, infoListCapnp.Len())

	for i := 0; i < infoListCapnp.Len(); i++ {
		serviceInfoCapnp := infoListCapnp.At(i)
		serviceInfoPOGS := UnmarshalServiceInfo(serviceInfoCapnp)
		infoListPOGS[i] = serviceInfoPOGS
	}

	return infoListPOGS
}
