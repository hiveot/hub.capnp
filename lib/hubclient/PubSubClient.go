// Package hubclient with helper functions for clients of the Hub
// This provides a resilient connection to the pubsub capabilities.
package hubclient

//// GetDevicePubSubClient obtains a device pubsub client
//func GetDevicePubSubClient(conn net.Conn, deviceID string) (pubsub.IDevicePubSub, error) {
//	ctx := context.Background()
//	cl := capnpclient.NewPubSubCapnpClient(ctx, conn)
//	deviceClient, err := cl.CapDevicePubSub(ctx, deviceID)
//	return deviceClient, err
//}
//
//// GetServicePubSubClient obtains a service pubsub client
//func GetServicePubSubClient(conn net.Conn, serviceID string) (pubsub.IServicePubSub, error) {
//	ctx := context.Background()
//	cl := capnpclient.NewPubSubCapnpClient(ctx, conn)
//	serviceClient, err := cl.CapServicePubSub(ctx, serviceID)
//	return serviceClient, err
//}
//
//// GetUserPubSubClient obtains a pubsub client for end-users
//func GetUserPubSubClient(conn net.Conn, userID string) (pubsub.IUserPubSub, error) {
//	ctx := context.Background()
//	cl := capnpclient.NewPubSubCapnpClient(ctx, conn)
//	userClient, err := cl.CapUserPubSub(ctx, userID)
//	return userClient, err
//}
