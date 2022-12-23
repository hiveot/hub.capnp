package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capserializer"
	"github.com/hiveot/hub/pkg/resolver/client"
)

// LauncherCapnpServer provides the capnproto RPC server for the service launcher
// This implements the capnproto generated interface Launcher_Server
// See hub.capnp/go/hubapi/launcher.capnp.go for the interface.
type LauncherCapnpServer struct {
	capRegSrv *client.CapRegistrationServer
	svc       launcher.ILauncher
}

func (capsrv *LauncherCapnpServer) List(ctx context.Context, call hubapi.CapLauncher_list) error {
	args := call.Args()
	onlyRunning := args.OnlyRunning()
	infoList, err := capsrv.svc.List(ctx, onlyRunning)
	if err == nil {
		res, _ := call.AllocResults()
		svcInfoListCapnp := capserializer.MarshalServiceInfoList(infoList)
		_ = res.SetInfoList(svcInfoListCapnp)
	}
	return err
}
func (capsrv *LauncherCapnpServer) StartService(ctx context.Context,
	call hubapi.CapLauncher_startService) error {
	args := call.Args()
	serviceName, _ := args.Name()
	serviceInfo, err := capsrv.svc.StartService(ctx, serviceName)
	res, _ := call.AllocResults()
	svcInfoCapnp := capserializer.MarshalServiceInfo(serviceInfo)
	_ = res.SetInfo(svcInfoCapnp)
	return err
}

func (capsrv *LauncherCapnpServer) StartAll(ctx context.Context, call hubapi.CapLauncher_startAll) error {
	err := capsrv.svc.StartAll(ctx)
	_, _ = call.AllocResults()
	return err
}
func (capsrv *LauncherCapnpServer) StopService(ctx context.Context, call hubapi.CapLauncher_stopService) error {
	args := call.Args()
	serviceName, _ := args.Name()
	serviceInfo, err := capsrv.svc.StopService(ctx, serviceName)
	res, _ := call.AllocResults()
	svcInfoCapnp := capserializer.MarshalServiceInfo(serviceInfo)
	_ = res.SetInfo(svcInfoCapnp)
	return err
}

func (capsrv *LauncherCapnpServer) StopAll(ctx context.Context, call hubapi.CapLauncher_stopAll) error {
	err := capsrv.svc.StopAll(ctx)
	_ = call
	return err
}

// StartLauncherCapnpServer starts the capnp server for the launcher service
//
//	lis is the socket server from whom to accept connections
//	svc is the instance of the launcher service
func StartLauncherCapnpServer(lis net.Listener, svc launcher.ILauncher) error {

	logrus.Infof("Starting launcher service capnp adapter on: %s", lis.Addr())
	capsrv := &LauncherCapnpServer{
		svc: svc,
	}
	capRegSrv := client.NewCapRegistrationServer(
		launcher.ServiceName,
		hubapi.CapLauncher_Methods(nil, capsrv))
	// the launcher does not have any exported capabilities (yet)
	//capRegSrv.ExportCapability("", []string{hubapi.ClientTypeService})

	main := hubapi.CapLauncher_ServerToClient(&LauncherCapnpServer{
		svc:       svc,
		capRegSrv: capRegSrv,
	})
	return caphelp.Serve(lis, capnp.Client(main), nil)
}
