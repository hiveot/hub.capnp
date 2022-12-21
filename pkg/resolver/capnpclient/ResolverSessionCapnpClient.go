package capnpclient

import (
	"context"
	"fmt"
	"net"
	"time"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

// ResolverSessionCapnpClient implements the IResolverSession interface
type ResolverSessionCapnpClient struct {
	rpcConn    *rpc.Conn                 // connection to capnp server
	capSession hubapi.CapResolverSession // capnp client of the service
}

// GetCapability obtains the capability with the given name.
// The caller must release the capability when done.
func (cl *ResolverSessionCapnpClient) GetCapability(ctx context.Context,
	clientID string, clientType string, capabilityName string, args []string) (
	capability capnp.Client, err error) {

	method, release := cl.capSession.GetCapability(ctx,
		func(params hubapi.CapResolverSession_getCapability_Params) error {
			_ = params.SetClientID(clientID)
			_ = params.SetClientType(clientType)
			_ = params.SetCapName(capabilityName)
			if args != nil {
				err = params.SetArgs(caphelp.MarshalStringList(args))
			}
			return err
		})
	defer release()
	// return a future. Caller must release
	// this does not detect a broken connection until the capability is used
	capability = method.Capability().AddRef()
	return capability, err
}

// ListCapabilities lists the available capabilities of the service
// Returns a list of capabilities that can be obtained through the service
func (cl *ResolverSessionCapnpClient) ListCapabilities(
	ctx context.Context) (infoList []resolver.CapabilityInfo, err error) {

	infoList = make([]resolver.CapabilityInfo, 0)
	method, release := cl.capSession.ListCapabilities(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		infoListCapnp, err2 := resp.InfoList()
		if err = err2; err == nil {
			infoList = capserializer.UnmarshalCapabilyInfoList(infoListCapnp)
		}
	}
	return infoList, err
}

// RegisterCapabilities registers a service's capabilities along with the CapProvider
func (cl *ResolverSessionCapnpClient) RegisterCapabilities(ctx context.Context,
	serviceID string, capInfoList []resolver.CapabilityInfo,
	capProvider hubapi.CapProvider) (err error) {

	capInfoListCapnp := capserializer.MarshalCapabilityInfoList(capInfoList)
	method, release := cl.capSession.RegisterCapabilities(ctx,
		func(params hubapi.CapResolverSession_registerCapabilities_Params) error {
			err = params.SetCapInfo(capInfoListCapnp)
			_ = params.SetServiceID(serviceID)
			_ = params.SetProvider(capProvider.AddRef()) // don't forget AddRef
			return err
		})
	defer release()
	_, err = method.Struct()
	return err
}

// Release this client and close the connection
// If capabilities provided by this client are still in use then do not release the client
// as the connection is used by the provided capabilities.
func (cl *ResolverSessionCapnpClient) Release() {
	cl.capSession.Release()
	if cl.rpcConn != nil {
		err := cl.rpcConn.Close()
		if err != nil {
			logrus.Warning(err)
		}
	}
}

// NewResolverSessionCapnpClient create a new resolver client for obtaining capnp capabilities.
//
// The provided connection is optional and intended for testing or running multiple resolvers.
// It is taken over by the client and closed when the client is released.
// In most cases simply pass nil to use the standard socket path.
//
//	conn is the optional network connection interface to use. nil to auto resolve.
func NewResolverSessionCapnpClient(ctx context.Context, conn net.Conn) (cl *ResolverSessionCapnpClient, err error) {

	// if no connection is provided use the local socket to connect to the service
	if conn == nil {
		conn, err = net.DialTimeout("unix", resolver.DefaultResolverPath, time.Second)
		if err != nil {
			logrus.Errorf("unable to connect to the resolver: %s", err)
			return nil, err
		}
	}
	transport := rpc.NewStreamTransport(conn)
	rpcConn := rpc.NewConn(transport, nil)
	capSession := hubapi.CapResolverSession(rpcConn.Bootstrap(ctx))

	cl = &ResolverSessionCapnpClient{
		rpcConn:    rpcConn,
		capSession: capSession,
	}
	ctx2, _ := context.WithTimeout(ctx, time.Second*3)
	err = capSession.Resolve(ctx2)
	if err != nil || !capSession.IsValid() {
		err = fmt.Errorf("Failed establishing RPC connecting: %s", err)
	}
	return cl, err
}
