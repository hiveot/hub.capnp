package cmd

import (
	"context"
	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/certs/capnpclient"
	"github.com/hiveot/hub/pkg/gateway/config"
	"github.com/hiveot/hub/pkg/mqttgw/service"
	"github.com/sirupsen/logrus"
	"os"
)

const serviceName = "mqttgw"
const mqttTcpPort = 8883
const mqttWsPort = 8884

// main launches the mqttgw gateway service using TLS websocket
func main() {
	var err error
	var ipAddr string

	f, _, _ := svcconfig.SetupFolderConfig(serviceName)
	cfg := config.NewGatewayConfig()
	_ = f.LoadConfig(&cfg)
	ipAddr = cfg.Address
	if ipAddr == "" {
		// listen on all addresses
		// ipAddr = listener.GetOutboundIP("").String()
	}

	// certificates are needed to serve TLS websocket connections
	// on each restart a new set of keys is used and a new certificate is requested.
	keys := certsclient.CreateECDSAKeys()
	serverCert, caCert, err := capnpclient.RenewServiceCert(serviceName, ipAddr, keys, f.Run)
	if err != nil {
		logrus.Panicf("certs service not reachable when starting the gateway: %s", err)
	}

	mqttService := service.NewMqttGatewayService()
	go mqttService.Start(mqttTcpPort, mqttWsPort, serverCert, caCert)

	// wait for the shutdown signal and close down
	listener.WaitForSignal(context.Background())
	err = mqttService.Stop()

	if err == nil {
		logrus.Warningf("%s service has shutdown gracefully", serviceName)
		os.Exit(0)
	} else {
		logrus.Errorf("%s service shutdown with error: %s", serviceName, err)
		os.Exit(-1)
	}
}
