package capnpserver

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

// ResolverSessionCapnpServer implements the capnp server of the resolver session
// This implements the capnp hubapi.CapResolverSession_server interface.
type ResolverSessionCapnpServer struct {
	session resolver.IResolverSession
}

// GetCapability invokes the requested method to return the capability it provides
func (capsrv *ResolverSessionCapnpServer) GetCapability(
	ctx context.Context, call hubapi.CapResolverSession_getCapability) (err error) {

	args := call.Args()
	capabilityName, _ := args.CapName()
	clientID, _ := args.ClientID()
	clientType, _ := args.ClientType()
	methodArgsCapnp, _ := args.Args()
	methodArgs := caphelp.UnmarshalStringList(methodArgsCapnp)
	logrus.Infof("get capability '%s'", capabilityName)
	capability, err := capsrv.session.GetCapability(ctx, clientID, clientType, capabilityName, methodArgs)
	if err != nil {
		return err
	} else if !capability.IsValid() {
		err = fmt.Errorf("Invalid capability returned for '%s'", capabilityName)
		logrus.Error(err)
		return err
	}
	// todo, can State be used, like for counting or other aspects?
	//capability.State().Metadata.Put("a", "b")
	resp, err := call.AllocResults()
	if err == nil {
		err = resp.SetCapability(capability)
	}
	if err != nil {
		logrus.Errorf("get capabilities failed: %s", err)
	}
	return err
}

// ListCapabilities returns the aggregated list of capabilities from all connected services.
func (capsrv *ResolverSessionCapnpServer) ListCapabilities(
	ctx context.Context, call hubapi.CapResolverSession_listCapabilities) (err error) {

	//logrus.Info("list capabilities")
	infoList, err := capsrv.session.ListCapabilities(ctx)
	resp, err2 := call.AllocResults()
	if err = err2; err == nil {
		infoListCapnp := capserializer.MarshalCapabilityInfoList(infoList)
		err = resp.SetInfoList(infoListCapnp)
	}
	return err
}

// RegisterCapabilities sets the capabilities from a provider
func (capsrv *ResolverSessionCapnpServer) RegisterCapabilities(
	ctx context.Context, call hubapi.CapResolverSession_registerCapabilities) (err error) {
	logrus.Info("registering capabilities")

	args := call.Args()
	providerID, _ := args.ServiceID()
	capInfoCapnp, _ := args.CapInfo()
	capInfo := capserializer.UnmarshalCapabilyInfoList(capInfoCapnp)
	capProvider := args.Provider()

	// use AddRef as capProvider would otherwise be released on exit
	err = capsrv.session.RegisterCapabilities(ctx, providerID, capInfo, capProvider.AddRef())
	return err
}

// NewResolverSessionCapnpServer creates a capnp server session to serve a new connection
func NewResolverSessionCapnpServer(session resolver.IResolverSession) *ResolverSessionCapnpServer {

	srv := &ResolverSessionCapnpServer{
		session: session,
	}
	return srv
}
