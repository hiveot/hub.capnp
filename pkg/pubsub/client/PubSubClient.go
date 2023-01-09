package client

import (
	"context"
	"net"

	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/capnpclient"
)

// GetDevicePubSubClient obtains a device pubsub client
func GetDevicePubSubClient(conn net.Conn, deviceID string) (pubsub.IDevicePubSub, error) {
	ctx := context.Background()
	cl := capnpclient.NewPubSubCapnpClient(ctx, conn)
	deviceClient, err := cl.CapDevicePubSub(ctx, deviceID)
	return deviceClient, err
}
