package capnpserver

import (
	"context"
	"net"

	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/pubsub"
)

// PubSubCapnpServer provides the capnp RPC server for pubsub services.
// This implements the capnproto generated interface PubSubService_Server
// See hub.capnp/go/hubapi/PubSubService.capnp.go for the interface.
type PubSubCapnpServer struct {
	svc pubsub.IPubSubService
}

// CapDevicePubSub provides the capability to pub/sub thing information as an IoT device.
func (capsrv *PubSubCapnpServer) CapDevicePubSub(
	ctx context.Context, call hubapi.CapPubSubService_capDevicePubSub) error {
	args := call.Args()
	deviceID, _ := args.DeviceID()
	deviceSvc := capsrv.svc.CapDevicePubSub(ctx, deviceID)

	capDeviceSvc := NewDevicePubSubCapnpServer(deviceSvc)
	capability := hubapi.CapDevicePubSub_ServerToClient(capDeviceSvc)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *PubSubCapnpServer) CapServicePubSub(
	ctx context.Context, call hubapi.CapPubSubService_capServicePubSub) error {

	args := call.Args()
	serviceID, _ := args.ServiceID()
	serviceSvc := capsrv.svc.CapServicePubSub(ctx, serviceID)

	capServiceSvc := NewServicePubSubCapnpServer(serviceSvc)
	capability := hubapi.CapServicePubSub_ServerToClient(capServiceSvc)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

func (capsrv *PubSubCapnpServer) CapUserPubSub(
	ctx context.Context, call hubapi.CapPubSubService_capUserPubSub) error {

	args := call.Args()
	userID, _ := args.UserID()
	userSvc := capsrv.svc.CapUserPubSub(ctx, userID)

	capServiceSvc := NewUserPubSubCapnpServer(userSvc)
	capability := hubapi.CapUserPubSub_ServerToClient(capServiceSvc)
	res, err := call.AllocResults()
	if err == nil {
		err = res.SetCap(capability)
	}
	return err
}

//
//// Release the service and free its resources
//func (srv *PubSubCapnpServer) Release() error {
//	return srv.svc.Release()
//}

func StartPubSubCapnpServer(
	ctx context.Context, lis net.Listener, svc pubsub.IPubSubService) error {

	logrus.Infof("Starting pubsub service capnp adapter on: %s", lis.Addr())
	main := hubapi.CapPubSubService_ServerToClient(&PubSubCapnpServer{
		svc: svc,
	})

	err := caphelp.CapServe(ctx, pubsub.ServiceName, lis, capnp.Client(main))
	return err
}
