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
)

// LauncherCapnpServer provides the capnproto RPC server for the service launcher
// This implements the capnproto generated interface Launcher_Server
// See hub.capnp/go/hubapi/launcher.capnp.go for the interface.
type LauncherCapnpServer struct {
	caphelp.HiveOTServiceCapnpServer
	pogo launcher.ILauncher
}

func (capsrv *LauncherCapnpServer) List(ctx context.Context, call hubapi.CapLauncher_list) error {
	args := call.Args()
	onlyRunning := args.OnlyRunning()
	infoList, err := capsrv.pogo.List(ctx, onlyRunning)
	if err == nil {
		res, _ := call.AllocResults()
		svcInfoListCapnp := capserializer.MarshalServiceInfoList(infoList)
		_ = res.SetInfoList(svcInfoListCapnp)
	}
	return err
}
func (capsrv *LauncherCapnpServer) Start(ctx context.Context, call hubapi.CapLauncher_start) error {
	args := call.Args()
	serviceName, _ := args.Name()
	serviceInfo, err := capsrv.pogo.Start(ctx, serviceName)
	res, _ := call.AllocResults()
	svcInfoCapnp := capserializer.MarshalServiceInfo(serviceInfo)
	_ = res.SetInfo(svcInfoCapnp)
	return err
}
func (capsrv *LauncherCapnpServer) Stop(ctx context.Context, call hubapi.CapLauncher_stop) error {
	args := call.Args()
	serviceName, _ := args.Name()
	serviceInfo, err := capsrv.pogo.Stop(ctx, serviceName)
	res, _ := call.AllocResults()
	svcInfoCapnp := capserializer.MarshalServiceInfo(serviceInfo)
	_ = res.SetInfo(svcInfoCapnp)
	return err
}

func (capsrv *LauncherCapnpServer) StopAll(ctx context.Context, call hubapi.CapLauncher_stopAll) error {
	err := capsrv.pogo.StopAll(ctx)
	_ = call
	return err
}

// StartLauncherCapnpServer starts the capnp server for the launcher service
//
//	ctx is the context for serving capabilities
//	lis is the socket server from whom to accept connections
//	srv is the instance of the launcher service
//	lc holds the launcher configuration
func StartLauncherCapnpServer(
	ctx context.Context, lis net.Listener, srv launcher.ILauncher) error {

	logrus.Infof("Starting launcher service capnp adapter on: %s", lis.Addr())

	main := hubapi.CapLauncher_ServerToClient(&LauncherCapnpServer{
		pogo: srv,
	})
	return caphelp.Serve(lis, capnp.Client(main))
}
