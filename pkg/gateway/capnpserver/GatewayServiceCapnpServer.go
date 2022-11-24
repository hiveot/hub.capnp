package capnpserver

import (
	"context"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpclient"
)

// GatewayServiceCapnpServer implements the capnp server of the gateway service
// This server does something special for the GetCapability method.
// Instead of calling the POGS service, it connects to the requested service directly, obtains the
// capnp capability and returns it to the client. This is a pure capnp 'capability' (pun intended)
// This implements the capnp hubapi.CapGatewayService interface.
type GatewayServiceCapnpServer struct {
	svc          gateway.IGatewayService
	socketFolder string
}

// GetCapability returns capability of the given service
// This is implemented here, rather than in the main service, as both incoming and outgoing requests
// are using capnproto. The shortest route is to obtain and provide it from here.
func (srv *GatewayServiceCapnpServer) GetCapability(
	ctx context.Context, call hubapi.CapGatewayService_getCapability) error {
	args := call.Args()
	serviceName, _ := args.Service()
	clientType, _ := args.ClientType()
	_ = clientType
	// TODO: obtain client ID (userID, serviceName, IoT deviceID
	// TODO: determine authentication status, cert authenticated or token authenticated
	// TODO: validate required clientType for this capability
	// TODO: re-use connections
	// TODO: free connections when done
	// TODO: invoke method to get another capability
	// TODO: pass parameeters to invoking method
	err := fmt.Errorf("dbd")
	//go func() {
	newctx := context.Background()
	rpcConn, capability, err := capnpclient.GetLocalCapability(newctx, srv.socketFolder, serviceName)
	_ = rpcConn
	if err == nil {
		res, err2 := call.AllocResults()
		if err2 == nil {
			err = res.SetCap(capability)
		}
	}
	//}()
	return err
}

func (srv *GatewayServiceCapnpServer) GetGatewayInfo(
	ctx context.Context, call hubapi.CapGatewayService_getGatewayInfo) error {

	gwInfo, err := srv.svc.GetGatewayInfo(ctx)
	res, err2 := call.AllocResults()

	if err == nil && err2 == nil {
		//gwInfoCapnp := MarshalGatewayInfo(gwInfo)
		// Marshal the response object
		_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		capListCapnp, _ := hubapi.NewCapabilityInfo_List(seg, int32(len(gwInfo.Capabilities)))
		for i, capInfo := range gwInfo.Capabilities {
			_, seg, _ = capnp.NewMessage(capnp.SingleSegment(nil))
			capInfoCapnp, _ := hubapi.NewCapabilityInfo(seg)
			_ = capInfoCapnp.SetName(capInfo.Name)
			_ = capInfoCapnp.SetService(capInfo.Service)
			_ = capListCapnp.Set(i, capInfoCapnp)
		}

		_, seg, _ = capnp.NewMessage(capnp.SingleSegment(nil))
		gwInfoCapnp, _ := hubapi.NewGatewayInfo(seg)
		gwInfoCapnp.SetLatency(int32(gwInfo.Latency))
		_ = gwInfoCapnp.SetUrl(gwInfo.URL)
		_ = gwInfoCapnp.SetCapabilities(capListCapnp)

		err = res.SetInfo(gwInfoCapnp)
	}
	return err
}

func (srv *GatewayServiceCapnpServer) Ping(
	ctx context.Context, call hubapi.CapGatewayService_ping) error {

	response, err := srv.svc.Ping(ctx)
	if err != nil {
		err = fmt.Errorf("ping somehow managed to fail")
		logrus.Error(err)
		return err
	}
	res, err := call.AllocResults()

	if err == nil {
		err = res.SetResponse(response)
	}
	return err
}

func StartGatewayServiceCapnpServer(
	ctx context.Context, lis net.Listener, svc gateway.IGatewayService, socketFolder string) error {
	srv := &GatewayServiceCapnpServer{
		svc:          svc,
		socketFolder: socketFolder,
	}
	main := hubapi.CapGatewayService_ServerToClient(srv)
	err := caphelp.CapServe(ctx, gateway.ServiceName, lis, capnp.Client(main))
	return err
}
