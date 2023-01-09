package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/server"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/lib/caphelp"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

// ResolverServiceCapnpServer implements the capnp server of the resolver service
// This implements the capnp hubapi.CapResolverService_server interface.
type ResolverServiceCapnpServer struct {
	service resolver.IResolverService
}

// GetCapability invokes the requested method to return the capability it provides
//func (capsrv *ResolverServiceCapnpServer) GetCapability(
//	ctx context.Context, call hubapi.CapResolverService_getCapability) (err error) {
//
//	args := call.Args()
//	capabilityName, _ := args.CapName()
//	clientID, _ := args.ClientID()
//	clientType, _ := args.ClientType()
//	methodArgsCapnp, _ := args.Args()
//	methodArgs := caphelp.UnmarshalStringList(methodArgsCapnp)
//	logrus.Infof("get capability '%s'", capabilityName)
//	capability, err := capsrv.service.GetCapability(ctx, clientID, clientType, capabilityName, methodArgs)
//	if err != nil {
//		return err
//	} else if !capability.IsValid() {
//		err = fmt.Errorf("Invalid capability returned for '%s'", capabilityName)
//		logrus.Error(err)
//		return err
//	}
//	// todo, can State be used, like for counting or other aspects?
//	//capability.State().Metadata.Put("a", "b")
//	resp, err := call.AllocResults()
//	if err == nil {
//		err = resp.SetCapability(capability)
//	}
//	if err != nil {
//		logrus.Errorf("get capabilities failed: %s", err)
//	}
//	return err
//}

// ListCapabilities returns the aggregated list of capabilities from all connected services.
func (capsrv *ResolverServiceCapnpServer) ListCapabilities(
	ctx context.Context, call hubapi.CapProvider_listCapabilities) (err error) {

	clientType, _ := call.Args().ClientType()
	infoList, err := capsrv.service.ListCapabilities(ctx, clientType)
	resp, err2 := call.AllocResults()
	if err = err2; err == nil {
		infoListCapnp := capserializer.MarshalCapabilityInfoList(infoList)
		err = resp.SetInfoList(infoListCapnp)
	}
	return err
}

//func (capsrv *ResolverServiceCapnpServer) HandleUnknownMethod(
//	ctx context.Context, r capnp.Recv) *server.Method {
//
//	// search the list of discovered capabilities for a method that matches the interfaceID and methodID
//	pc := capsrv.service.HandleUnknownMethod(ctx, r)
//	if pc == nil {
//		r.Reject(capnp.Unimplemented("unimplemented"))
//		return nil
//	}
//	return pc
//}

// RegisterCapabilities sets the capabilities from a provider
//func (capsrv *ResolverServiceCapnpServer) RegisterCapabilities(
//	ctx context.Context, call hubapi.CapResolverService_registerCapabilities) (err error) {
//	logrus.Info("registering capabilities")
//
//	args := call.Args()
//	providerID, _ := args.ServiceID()
//	capInfoCapnp, _ := args.CapInfo()
//	capInfo := capserializer.UnmarshalCapabilyInfoList(capInfoCapnp)
//	capProvider := args.Provider()
//
//	// use AddRef as capProvider would otherwise be released on exit
//	err = capsrv.service.RegisterCapabilities(ctx, providerID, capInfo, capProvider.AddRef())
//	return err
//}

// NewResolverServiceCapnpServer creates a capnp server service to serve a new connection
//func NewResolverServiceCapnpServer(service resolver.IResolverService) *ResolverServiceCapnpServer {
//
//	srv := &ResolverServiceCapnpServer{
//		service: service,
//	}
//	// map capnp calls to service calls
//	//main := hubapi.CapResolverService_ServerToClient(srv)
//	//err := caphelp.Serve(lis, capnp.Client(main), nil)
//	return srv
//}

func StartResolverServiceCapnpServer(
	service resolver.IResolverService, lis net.Listener,
	handleUnknownMethod func(m capnp.Method) *server.Method) {

	srv := &ResolverServiceCapnpServer{
		service: service,
	}

	//main := hubapi.CapResolverService_ServerToClient(srv)
	c, _ := hubapi.CapResolverService_Server(srv).(server.Shutdowner)
	methods := hubapi.CapResolverService_Methods(nil, srv)
	clientHook := server.New(methods, srv, c)
	clientHook.HandleUnknownMethod = handleUnknownMethod

	//resServer := hubapi.CapResolverService_NewServer(s)
	resClient := capnp.NewClient(clientHook)
	main := hubapi.CapResolverService(resClient)

	_ = caphelp.Serve(resolver.ServiceName, lis, capnp.Client(main), nil)

}
