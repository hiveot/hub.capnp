package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/provisioning"
	"github.com/hiveot/hub/pkg/provisioning/capnp4POGS"
)

// ManageProvisioningCapnpServer provides the capnproto RPC server for IOT device provisioning.
// This implements the capnproto generated interface CapManageProvisioning_Server
type ManageProvisioningCapnpServer struct {
	srv provisioning.IManageProvisioning
}

func (capsrv *ManageProvisioningCapnpServer) AddOOBSecrets(
	ctx context.Context, call hubapi.CapManageProvisioning_addOOBSecrets) error {

	args := call.Args()
	secretsCapnp, _ := args.OobSecrets()
	secretsPOGS := capnp4POGS.OobSecretsCapnp2POGS(secretsCapnp)
	err := capsrv.srv.AddOOBSecrets(ctx, secretsPOGS)
	return err
}

func (capsrv *ManageProvisioningCapnpServer) ApproveRequest(
	ctx context.Context, call hubapi.CapManageProvisioning_approveRequest) error {

	args := call.Args()
	deviceID, _ := args.DeviceID()
	err := capsrv.srv.ApproveRequest(ctx, deviceID)
	return err
}

func (capsrv *ManageProvisioningCapnpServer) GetApprovedRequests(
	ctx context.Context, call hubapi.CapManageProvisioning_getApprovedRequests) error {

	statusList, err := capsrv.srv.GetApprovedRequests(ctx)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		if err2 == nil {
			statusListCapnp := capnp4POGS.ProvStatusListPOGS2Capnp(statusList)
			res.SetRequests(statusListCapnp)
		}
	}
	return err
}

func (capsrv *ManageProvisioningCapnpServer) GetPendingRequests(
	ctx context.Context, call hubapi.CapManageProvisioning_getPendingRequests) error {

	statusList, err := capsrv.srv.GetPendingRequests(ctx)
	if err == nil {
		res, err2 := call.AllocResults()
		err = err2
		if err2 == nil {
			statusListCapnp := capnp4POGS.ProvStatusListPOGS2Capnp(statusList)
			res.SetRequests(statusListCapnp)
		}
	}
	return err
}

func NewManageProvisioningCapnpServer(srv provisioning.IManageProvisioning) *ManageProvisioningCapnpServer {
	capsrv := &ManageProvisioningCapnpServer{srv: srv}
	return capsrv
}
