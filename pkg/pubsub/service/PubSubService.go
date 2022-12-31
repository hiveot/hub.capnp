package service

import (
	"context"

	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/core"
)

// PubSubService implements the publish/subscribe service
// This implements the IPubSubService interface
//
// This service main task is to issue capabilities to devices, services and end-users
type PubSubService struct {
	core *core.PubSubCore
}

// CapDevicePubSub provides the capability to pub/sub thing information as an IoT device.
// The issuer must only provide this capability after verifying the device ID.
func (svc *PubSubService) CapDevicePubSub(_ context.Context, deviceID string) (pubsub.IDevicePubSub, error) {
	devicePubSub := NewDevicePubSub(deviceID, svc.core)
	return devicePubSub, nil
}

// CapServicePubSub provides the capability to pub/sub thing information as a hub service.
// Hub services can publish their own information and receive events from any thing.
func (svc *PubSubService) CapServicePubSub(_ context.Context, serviceID string) (pubsub.IServicePubSub, error) {
	servicePubSub := NewServicePubSub(serviceID, svc.core)
	return servicePubSub, nil
}

// CapUserPubSub provides the capability for an end-user to publish or subscribe to messages.
// The caller must authenticate the user and provide appropriate configuration.
//
//	userID is the login ID of an authenticated user
func (svc *PubSubService) CapUserPubSub(_ context.Context, userID string) (pubsub.IUserPubSub, error) {
	userPubSub := NewUserPubSub(userID, svc.core)
	return userPubSub, nil
}

// Release the service and free its resources
//func (svc *PubSubService) Release() error {
//	err := svc.core.Stop()
//	return err
//}

func (svc *PubSubService) Start() error {
	err := svc.core.Start()
	return err
}
func (svc *PubSubService) Stop() error {
	err := svc.core.Stop()
	return err
}

// NewPubSubService creates a new instance of the pubsub
// returns an error if start fails
func NewPubSubService() *PubSubService {
	pubsubCore := core.NewPubSubCore()
	svc := &PubSubService{
		core: pubsubCore,
	}
	return svc
}
