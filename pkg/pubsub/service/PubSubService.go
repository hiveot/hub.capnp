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
func (svc *PubSubService) CapDevicePubSub(ctx context.Context, deviceID string) pubsub.IDevicePubSub {
	devicePubSub := NewDevicePubSub(deviceID, svc.core)
	return devicePubSub
}

// CapServicePubSub provides the capability to pub/sub thing information as a hub service.
// Hub services can publish their own information and receive events from any thing.
func (svc *PubSubService) CapServicePubSub(ctx context.Context, serviceID string) pubsub.IServicePubSub {
	servicePubSub := NewServicePubSub(serviceID, svc.core)
	return servicePubSub
}

// CapUserPubSub provides the capability for an end-user to publish or subscribe to messages.
// The caller must authenticate the user and provide appropriate configuration.
//
//	userID is the login ID of an authenticated user
func (svc *PubSubService) CapUserPubSub(ctx context.Context, userID string) (pub pubsub.IUserPubSub) {
	userPubSub := NewUserPubSub(userID, svc.core)
	return userPubSub
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
