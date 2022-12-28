package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capserializer"
)

// LauncherCapnpClient provides a POGS wrapper around the launcher capnp client
// This implements the ILauncher interface
type LauncherCapnpClient struct {
	connection *rpc.Conn          // connection to the capnp server
	capability hubapi.CapLauncher // capnp client to the server
}

// List services
func (cl *LauncherCapnpClient) List(ctx context.Context, onlyRunning bool) (infoList []launcher.ServiceInfo, err error) {

	method, release := cl.capability.List(ctx,
		func(params hubapi.CapLauncher_list_Params) error {
			params.SetOnlyRunning(onlyRunning)
			return nil
		})

	defer release()
	resp, err := method.Struct()
	if err == nil {
		infoListCapnp, _ := resp.InfoList()
		infoList = capserializer.UnmarshalServiceInfoList(infoListCapnp)
	}
	return infoList, err
}

// Release the capability
func (cl *LauncherCapnpClient) Release() {
	cl.capability.Release()
}

func (cl *LauncherCapnpClient) StartService(
	ctx context.Context, name string) (serviceInfo launcher.ServiceInfo, err error) {

	method, release := cl.capability.StartService(ctx,
		func(params hubapi.CapLauncher_startService_Params) error {
			err := params.SetName(name)
			return err
		})
	defer release()
	resp, err := method.Struct()
	serviceInfoCapnp, _ := resp.Info()
	serviceInfo = capserializer.UnmarshalServiceInfo(serviceInfoCapnp)
	return serviceInfo, err
}

// StartAll services
func (cl *LauncherCapnpClient) StartAll(ctx context.Context) (err error) {

	method, release := cl.capability.StartAll(ctx, nil)
	defer release()
	_, err = method.Struct()
	return err
}

func (cl *LauncherCapnpClient) StopService(
	ctx context.Context, name string) (serviceInfo launcher.ServiceInfo, err error) {

	method, release := cl.capability.StopService(ctx,
		func(params hubapi.CapLauncher_stopService_Params) error {
			err := params.SetName(name)
			return err
		})
	defer release()
	resp, err := method.Struct()
	serviceInfoCapnp, _ := resp.Info()
	serviceInfo = capserializer.UnmarshalServiceInfo(serviceInfoCapnp)
	return serviceInfo, err
}

// StopAll running services
func (cl *LauncherCapnpClient) StopAll(ctx context.Context) (err error) {

	method, release := cl.capability.StopAll(ctx, nil)
	defer release()
	_, err = method.Struct()
	return err
}

// NewLauncherCapnpClient returns a service client using the capnp protocol
//
//	ctx is the context for obtaining capabilities
//	connection is the connection to the capnp RPC server
func NewLauncherCapnpClient(ctx context.Context, connection net.Conn) (*LauncherCapnpClient, error) {
	var cl *LauncherCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapLauncher(rpcConn.Bootstrap(ctx))

	cl = &LauncherCapnpClient{
		connection: rpcConn,
		capability: capability,
	}

	return cl, nil
}
