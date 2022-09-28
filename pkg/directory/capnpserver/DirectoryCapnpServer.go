package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryCapnpServer provides the capnp RPC server for directory services
// This implements the capnproto generated interface Directory_Server
// See hub.capnp/go/hubapi/DirectoryStore.capnp.go for the interface.
type DirectoryCapnpServer struct {
	srv directory.IDirectory
}

func (capsrv *DirectoryCapnpServer) CapReadDirectory(
	ctx context.Context, call hubapi.CapDirectory_capReadDirectory) error {

	readCapSrv := NewReadDirectoyCapnpServer(capsrv.srv.CapReadDirectory())
	capability := hubapi.CapReadDirectory_ServerToClient(readCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *DirectoryCapnpServer) CapUpdateDirectory(
	ctx context.Context, call hubapi.CapDirectory_capUpdateDirectory) error {

	updateCapSrv := NewUpdateDirectoryCapnpServer(capsrv.srv.CapUpdateDirectory())
	capability := hubapi.CapUpdateDirectory_ServerToClient(updateCapSrv)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

// StartDirectoryCapnpServer starts the directory service capnp protocol server
func StartDirectoryCapnpServer(ctx context.Context, lis net.Listener, srv directory.IDirectory) error {

	logrus.Infof("Starting directory service capnp adapter on: %s", lis.Addr())

	main := hubapi.CapDirectory_ServerToClient(&DirectoryCapnpServer{
		srv: srv,
	})

	err := caphelp.CapServe(ctx, lis, capnp.Client(main))
	return err
}
