package capnp4POGS

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/launcher"
)

func ServiceInfoCapnp2POGS(serviceInfoCapnp hubapi.ServiceInfo) launcher.ServiceInfo {
	errorText, _ := serviceInfoCapnp.Error()
	modifiedTime, _ := serviceInfoCapnp.ModifiedTime()
	serviceName, _ := serviceInfoCapnp.Name()
	servicePath, _ := serviceInfoCapnp.Path()
	startTime, _ := serviceInfoCapnp.StartTime()
	stopTime, _ := serviceInfoCapnp.StopTime()

	serviceInfoPOGS := launcher.ServiceInfo{
		CPU:          int(serviceInfoCapnp.Cpu()),
		MEM:          int(serviceInfoCapnp.Mem()),
		Error:        errorText,
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

func ServiceInfoPOGS2Capnp(serviceInfoPOGS launcher.ServiceInfo) hubapi.ServiceInfo {
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	siCapnp, _ := hubapi.NewServiceInfo(seg)
	siCapnp.SetCpu(int32(serviceInfoPOGS.CPU))
	siCapnp.SetMem(int32(serviceInfoPOGS.MEM))
	_ = siCapnp.SetError(serviceInfoPOGS.Error)
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
