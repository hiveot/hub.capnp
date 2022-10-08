package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/provisioning"
	"github.com/hiveot/hub/pkg/provisioning/capnp4POGS"
)

// RequestProvisioningCapnpServer provides the capnproto RPC server to request device provisioning
type RequestProvisioningCapnpServer struct {
	pogosrv provisioning.IRequestProvisioning
}

func (capsrv *RequestProvisioningCapnpServer) SubmitProvisioningRequest(
	ctx context.Context, call hubapi.CapRequestProvisioning_submitProvisioningRequest) error {
	args := call.Args()
	deviceID, _ := args.DeviceID()
	pubKeyPEM, _ := args.PubKeyPEM()
	secretMd5, _ := args.Md5Secret()
	status, err := capsrv.pogosrv.SubmitProvisioningRequest(ctx, deviceID, secretMd5, pubKeyPEM)
	if err == nil {
		res, _ := call.AllocResults()
		provStatusCapnp := capnp4POGS.ProvStatusPOGS2Capnp(status)
		_ = res.SetStatus(provStatusCapnp)
	}
	return err
}
