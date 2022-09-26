package service

import (
	"context"
	"fmt"
	"net"

	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub/pkg/provisioning/service/oobprovserver"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
)

// ProvisioningCapnpAdapter is a capnproto adapter for the provisioning services.
// This implements the capnproto generated interface ProvisioningService_Server
// See hub.capnp/go/hubapi/Provisioning.capnp.go for the interface.
type ProvisioningCapnpAdapter struct {
	svc *oobprovserver.OobProvServer
}

func (adpt *ProvisioningCapnpAdapter) AddOOBSecret(
	context.Context, hubapi.ProvisioningService_addOOBSecret) error {
	return fmt.Errorf("Not implemented")
}

func (adpt *ProvisioningCapnpAdapter) ApproveRequest(
	context.Context, hubapi.ProvisioningService_approveRequest) error {
	return fmt.Errorf("Not implemented")
}

func (adpt *ProvisioningCapnpAdapter) GetPendingRequests(
	context.Context, hubapi.ProvisioningService_getPendingRequests) error {
	return fmt.Errorf("Not implemented")
}

func (adpt *ProvisioningCapnpAdapter) RefreshProvisioning(
	context.Context, hubapi.ProvisioningService_refreshProvisioning) error {
	return fmt.Errorf("Not implemented")
}

func (adpt *ProvisioningCapnpAdapter) SubmitProvisioningRequest(
	context.Context, hubapi.ProvisioningService_submitProvisioningRequest) error {
	return fmt.Errorf("Not implemented")
}

// StartProvisioningCapnpAdapter starts the provisioning service capnp protocol server
func StartProvisioningCapnpAdapter(
	ctx context.Context, lis net.Listener, service *oobprovserver.OobProvServer) error {

	main := hubapi.ProvisioningService_ServerToClient(&ProvisioningCapnpAdapter{
		svc: service,
	})

	return caphelp.CapServe(ctx, lis, capnp.Client(main))
}
