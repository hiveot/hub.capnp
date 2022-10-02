package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/provisioning"
	"github.com/hiveot/hub/pkg/provisioning/capnp4POGS"
)

// RequestProvisioningCapnpClient provides the POGS interface with the capability to send provisioning requests
type RequestProvisioningCapnpClient struct {
	// The capnp client
	capability hubapi.CapRequestProvisioning
}

// SubmitProvisioningRequest passes the provisioning request to the server via the capnp protocol
func (cl *RequestProvisioningCapnpClient) SubmitProvisioningRequest(
	ctx context.Context, deviceID string, md5Secret string, pubKeyPEM string) (
	provStatus provisioning.ProvisionStatus, err error) {

	method, release := cl.capability.SubmitProvisioningRequest(ctx,
		func(params hubapi.CapRequestProvisioning_submitProvisioningRequest_Params) error {
			err2 := params.SetDeviceID(deviceID)
			_ = params.SetMd5Secret(md5Secret)
			_ = params.SetPubKeyPEM(pubKeyPEM)
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		provStatusCapnp, err2 := resp.Status()
		err = err2
		provStatus = capnp4POGS.ProvStatusCapnp2POGS(provStatusCapnp)
	}
	return provStatus, err
}

// NewRequestProvisioningCapnpClient returns an instance of the POGS wrapper around the capnp api
func NewRequestProvisioningCapnpClient(cap hubapi.CapRequestProvisioning) *RequestProvisioningCapnpClient {
	cl := &RequestProvisioningCapnpClient{capability: cap}
	return cl
}
