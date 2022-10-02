package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/provisioning"
	"github.com/hiveot/hub/pkg/provisioning/capnp4POGS"
)

// RefreshProvisioningCapnpServer provides the capnproto RPC server to Refresh device provisioning
type RefreshProvisioningCapnpServer struct {
	srv provisioning.IRefreshProvisioning
}

func (capsrv *RefreshProvisioningCapnpServer) RefreshDeviceCert(
	ctx context.Context, call hubapi.CapRefreshProvisioning_refreshDeviceCert) error {

	args := call.Args()
	//deviceID, _ := args.DeviceID()
	certPEM, _ := args.CertPEM()
	status, err := capsrv.srv.RefreshDeviceCert(ctx, certPEM)
	if err == nil {
		res, _ := call.AllocResults()
		provStatusCapnp := capnp4POGS.ProvStatusPOGS2Capnp(status)
		res.SetStatus(provStatusCapnp)
	}

	return err
}

// NewRefreshProvisioningCapnpServer serves refreshing a device certificate
func NewRefreshProvisioningCapnpServer(srv provisioning.IRefreshProvisioning) *RefreshProvisioningCapnpServer {
	capsrv := &RefreshProvisioningCapnpServer{srv: srv}

	return capsrv
}
