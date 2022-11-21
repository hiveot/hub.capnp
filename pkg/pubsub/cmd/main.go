package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/pubsub"
	"github.com/hiveot/hub/pkg/pubsub/capnpserver"
	"github.com/hiveot/hub/pkg/pubsub/service"
)

// Start the history store service
func main() {
	logging.SetLogging("info", "")
	ctx := context.Background()

	f := svcconfig.GetFolders("", false)
	//cfg := NewPubSubConfig(f.Stores)
	//f = svcconfig.LoadServiceConfig(launcher.ServiceName, false, &cfg)

	srvListener := listener.CreateServiceListener(f.Run, pubsub.ServiceName)

	svc, err := service.StartPubSubService()
	if err != nil {
		logrus.Panicf("unable launch the pubsub service: %s", err)
	}

	// connections via capnp RPC
	if err == nil {
		logrus.Infof("PubSubServiceCapnpServer starting on: %s", srvListener.Addr())
		_ = capnpserver.StartPubSubCapnpServer(ctx, srvListener, svc)
	}
	if err != nil {
		msg := fmt.Sprintf("ERROR: Service '%s' failed to start: %s\n", pubsub.ServiceName, err)
		logrus.Fatal(msg)
	}
	logrus.Infof("PubSub service ended gracefully")
	err = svc.Release()
	os.Exit(0)
}
