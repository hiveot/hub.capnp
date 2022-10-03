package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/provisioning"
	"github.com/hiveot/hub/pkg/provisioning/capnp4POGS"
)

// RefreshProvisioningCapnpClient provides the POGS interface with the capability to send provisioning requests
// This implements the RefreshDeviceCert interface
type RefreshProvisioningCapnpClient struct {
	// The capnp client
	capability hubapi.CapRefreshProvisioning
}

// RefreshDeviceCert passes the certificate refresh request to the server via the capnp protocol
func (cl *RefreshProvisioningCapnpClient) RefreshDeviceCert(
	ctx context.Context, certPEM string) (
	provStatus provisioning.ProvisionStatus, err error) {

	resp, release := cl.capability.RefreshDeviceCert(ctx,
		func(params hubapi.CapRefreshProvisioning_refreshDeviceCert_Params) error {
			err2 := params.SetCertPEM(certPEM)
			return err2
		})
	defer release()
	method, err := resp.Struct()
	if err == nil {
		statusCapnp, err2 := method.Status()
		err = err2
		provStatus = capnp4POGS.ProvStatusCapnp2POGS(statusCapnp)
	}
	return provStatus, err
}

// NewRefreshProvisioningCapnpClient returns an instance of the POGS wrapper around the capnp api
func NewRefreshProvisioningCapnpClient(cap hubapi.CapRefreshProvisioning) *RefreshProvisioningCapnpClient {
	cl := &RefreshProvisioningCapnpClient{capability: cap}
	return cl
}