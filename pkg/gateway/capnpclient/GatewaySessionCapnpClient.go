package capnpclient

import (
	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"
	"context"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capserializer"
)

type GatewaySessionCapnpClient struct {
	connection *rpc.Conn          // connection to capnp server
	capability hubapi.CapProvider // capnp client of the gateway session
}

// ListCapabilities lists the available capabilities of the service
// Returns a list of capabilities that can be obtained through the service
func (cl *GatewaySessionCapnpClient) ListCapabilities(
	ctx context.Context) (infoList []resolver.CapabilityInfo, err error) {

	infoList = make([]resolver.CapabilityInfo, 0)
	method, release := cl.capability.ListCapabilities(ctx, nil)
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

// Release the client
func (cl *GatewaySessionCapnpClient) Release() {
	cl.capability.Release()
	if cl.connection != nil {
		err := cl.connection.Close()
		if err != nil {
			logrus.Error(err)
		}
	}
}

// NewGatewaySessionCapnpClient returns a POGS wrapper around the gateway capnp instance
func NewGatewaySessionCapnpClient(capClient capnp.Client) resolver.ICapProvider {
	capGateway := hubapi.CapProvider(capClient)
	gws := GatewaySessionCapnpClient{
		capability: capGateway,
		connection: nil,
	}
	return &gws
}
