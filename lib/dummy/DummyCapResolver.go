package dummy

import (
	"context"

	"capnproto.org/go/capnp/v3"
	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/resolver"
)

// DummyCapResolver for testing. This implements ICapResolver and ICapProvider interfaces
type DummyCapResolver struct {
}

// --- Provider ---

func (dummy *DummyCapResolver) GetCapability(ctx context.Context,
	clientID string, clientType string, capName string, args []string) (capnp.Client, error) {
	cl := capnp.Client{}
	return cl, nil
}

func (dummy *DummyCapResolver) ListCapabilities(ctx context.Context, clientType string) (capInfo []resolver.CapabilityInfo, err error) {
	return nil, nil
}

// --- Resolver ---

func (dummy *DummyCapResolver) RegisterCapabilities(ctx context.Context, providerID string, capInfo []resolver.CapabilityInfo, provider *hubapi.CapProvider) error {
	return nil
}

func (dummy *DummyCapResolver) OnIncomingConnection(rpcConn *rpc.Conn) {

}

func NewDummyCapResolver() *DummyCapResolver {
	dummy := &DummyCapResolver{}
	return dummy
}