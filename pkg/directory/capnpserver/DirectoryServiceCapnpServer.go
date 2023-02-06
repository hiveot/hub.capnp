package capnpserver

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/resolver/capprovider"
)

// DirectoryServiceCapnpServer provides the capnp RPC server for directory services
// This implements the capnproto generated interface Directory_Server
// See hub.capnp/go/hubapi/DirectoryStore.capnp.go for the interface.
type DirectoryServiceCapnpServer struct {
	svc directory.IDirectory
}

func (capsrv *DirectoryServiceCapnpServer) CapReadDirectory(
	ctx context.Context, call hubapi.CapDirectoryService_capReadDirectory) error {
	clientID, _ := call.Args().ClientID()
	capReadDirectory, _ := capsrv.svc.CapReadDirectory(ctx, clientID)
	readCapSrv := &ReadDirectoryCapnpServer{
		srv: capReadDirectory,
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

	clientID, _ := call.Args().ClientID()
	capUpdateDirectory, _ := capsrv.svc.CapUpdateDirectory(ctx, clientID)
	updateCapSrv := &UpdateDirectoryCapnpServer{
		srv: capUpdateDirectory,
	}

	capability := hubapi.CapUpdateDirectory_ServerToClient(updateCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartDirectoryServiceCapnpServer starts the capnp protocol server for the directory service
func StartDirectoryServiceCapnpServer(svc directory.IDirectory, lis net.Listener) error {
	serviceName := directory.ServiceName

	srv := &DirectoryServiceCapnpServer{
		svc: svc,
	}
	// the provider serves the exported capabilities
	methods := hubapi.CapDirectoryService_Methods(nil, srv)
	capProv := capprovider.NewCapServer(serviceName, methods)

	capProv.ExportCapability(hubapi.CapNameReadDirectory,
		[]string{hubapi.ClientTypeUser, hubapi.ClientTypeService})

	capProv.ExportCapability(hubapi.CapNameUpdateDirectory,
		[]string{hubapi.ClientTypeService})

	logrus.Infof("Starting '%s' service capnp adapter listening on: %s", serviceName, lis.Addr())
	err := capProv.Start(lis)
	return err
}
