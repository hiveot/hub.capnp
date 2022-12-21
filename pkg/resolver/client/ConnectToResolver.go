package client

import (
	"context"
	"net"
	"time"

	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpclient"
)

// ConnectToResolver is a helper that starts a new session with the resolver
// Users should call Release when done. This will close the connection and any
// capabilities obtained from the resolver.
//
//	resolverSocket is the path to the socket the resolver listens on or "" for the default
//
// This returns the resolver client
func ConnectToResolver(resolverSocket string) (
	resolverClient *capnpclient.ResolverSessionCapnpClient, err error) {

	if resolverSocket == "" {
		resolverSocket = resolver.DefaultResolverPath
	}
	conn, err := net.DialTimeout("unix", resolverSocket, time.Second*10)
	ctx := context.Background()
	cl, err := capnpclient.NewResolverSessionCapnpClient(ctx, conn)
	return cl, err
}
