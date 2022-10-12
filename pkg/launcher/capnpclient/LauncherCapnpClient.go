package capnpclient

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capnp4POGS"
)

// LauncherCapnpClient provides a POGS wrapper around the launcher capnp client
// This implements the ILauncher interface
type LauncherCapnpClient struct {
	connection *rpc.Conn          // connection to the capnp server
	capability hubapi.CapLauncher // capnp client
	ctx        context.Context
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
		infoList = capnp4POGS.InfoListCapnp2POGS(infoListCapnp)
	}
	return infoList, err
}

// Start a service
func (cl *LauncherCapnpClient) Start(
	ctx context.Context, name string) (serviceInfo launcher.ServiceInfo, err error) {

	method, release := cl.capability.Start(ctx,
		func(params hubapi.CapLauncher_start_Params) error {
			err := params.SetName(name)
			return err
		})
	defer release()
	resp, err := method.Struct()
	serviceInfoCapnp, _ := resp.Info()
	serviceInfo = capnp4POGS.ServiceInfoCapnp2POGS(serviceInfoCapnp)
	return serviceInfo, err
}

// Stop a running service
func (cl *LauncherCapnpClient) Stop(
	ctx context.Context, name string) (serviceInfo launcher.ServiceInfo, err error) {

	method, release := cl.capability.Stop(ctx,
		func(params hubapi.CapLauncher_stop_Params) error {
			err := params.SetName(name)
			return err
		})
	defer release()
	resp, err := method.Struct()
	serviceInfoCapnp, _ := resp.Info()
	serviceInfo = capnp4POGS.ServiceInfoCapnp2POGS(serviceInfoCapnp)
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
//  ctx is the context for obtaining capabilities
//  connection is the connection to the capnp RPC server
func NewLauncherCapnpClient(ctx context.Context, connection net.Conn) (*LauncherCapnpClient, error) {
	var cl *LauncherCapnpClient
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	capability := hubapi.CapLauncher(rpcConn.Bootstrap(ctx))

	cl = &LauncherCapnpClient{
		connection: rpcConn,
		capability: capability,
		ctx:        ctx,
	}

	return cl, nil
}
