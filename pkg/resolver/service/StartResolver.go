package service

import (
	"context"
	"net"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpserver"
)

// StartResolver is a helper for starting the resolver service and its capnp server
// This returns a stop function or an error if start fails.
func StartResolver(socketPath string) (stopFn func(), err error) {
	if socketPath == "" {
		socketPath = resolver.DefaultResolverPath
	}
	_ = os.RemoveAll(socketPath)
	svc := NewResolverService()
	err = svc.Start(context.Background())
	if err != nil {
		logrus.Panicf("Failed to start with socket dir %s", socketPath)
	}

	lis, err := net.Listen("unix", socketPath)
	err = svc.Start(context.Background())
	if err == nil {
		go capnpserver.StartResolverCapnpServer(lis, svc)
	}
	return func() {
		_ = svc.Stop()
		_ = lis.Close()
	}, err
}
