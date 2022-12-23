package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/resolver/client"
)

// DirectoryServiceCapnpServer provides the capnp RPC server for directory services
// This implements the capnproto generated interface Directory_Server
// See hub.capnp/go/hubapi/DirectoryStore.capnp.go for the interface.
type DirectoryServiceCapnpServer struct {
	capRegSrv *client.CapRegistrationServer
	svc       directory.IDirectory
}

func (capsrv *DirectoryServiceCapnpServer) CapReadDirectory(
	ctx context.Context, call hubapi.CapDirectoryService_capReadDirectory) error {

	readCapSrv := &ReadDirectoryCapnpServer{
		srv: capsrv.svc.CapReadDirectory(ctx),
	}

	capability := hubapi.CapReadDirectory_ServerToClient(readCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *DirectoryServiceCapnpServer) CapUpdateDirectory(
	ctx context.Context, call hubapi.CapDirectoryService_capUpdateDirectory) error {

	updateCapSrv := &UpdateDirectoryCapnpServer{
		srv: capsrv.svc.CapUpdateDirectory(ctx),
	}

	capability := hubapi.CapUpdateDirectory_ServerToClient(updateCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartDirectoryServiceCapnpServer starts the capnp protocol server for the directory service
func StartDirectoryServiceCapnpServer(lis net.Listener, svc directory.IDirectory) error {

	logrus.Infof("Starting directory service capnp adapter on: %s", lis.Addr())

	srv := &DirectoryServiceCapnpServer{
		svc: svc,
	}
	capRegSrv := client.NewCapRegistrationServer(
		directory.ServiceName, hubapi.CapDirectoryService_Methods(nil, srv))
	srv.capRegSrv = capRegSrv
	capRegSrv.ExportCapability("capReadDirectory", []string{hubapi.ClientTypeUser, hubapi.ClientTypeService})
	capRegSrv.ExportCapability("capUpdateDirectory", []string{hubapi.ClientTypeService})

	main := hubapi.CapDirectoryService_ServerToClient(srv)
	err := rpc.Serve(lis, capnp.Client(main))
	return err
}
