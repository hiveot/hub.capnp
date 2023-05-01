package main

import (
	"context"
	"net"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"

	"github.com/hiveot/hub/pkg/certs/capnpclient"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/gateway/capnpserver"
	"github.com/hiveot/hub/pkg/gateway/config"
	"github.com/hiveot/hub/pkg/gateway/service"
	"github.com/hiveot/hub/pkg/resolver"
)

// main launches the gateway service using TLS socket
func main() {
	var serviceName = gateway.ServiceName
	var err error
	var svc *service.GatewayService
	var lisTcp net.Listener
	var lisWS net.Listener
	var ipAddr string

	f, _, _ := svcconfig.SetupFolderConfig(serviceName)
	cfg := config.NewGatewayConfig()
	_ = f.LoadConfig(&cfg)
	ipAddr = cfg.Address
	if ipAddr == "" {
		// listen on all addresses
		// ipAddr = listener.GetOutboundIP("").String()
	}

	// certificates are needed for TLS connections to the capnp server.
	// on each restart a new set of keys is used and a new certificate is requested.
	keys := certsclient.CreateECDSAKeys()
	serverCert, caCert, err := capnpclient.RenewServiceCert(serviceName, ipAddr, keys, f.Run)
	if err != nil {
		logrus.Panicf("certs service not reachable when starting the gateway: %s", err)
	}

	// Start the gateway service instance.
	if err == nil {
		// Need the resolver service socket to connect to. This is a fixed path.
		//resolverPath := path.Join(f.Run, resolver.ServiceName+".socket")
		resolverPath := resolver.DefaultResolverPath
		svc = service.NewGatewayService(resolverPath, nil)
	}
	err = svc.Start()
	if err != nil {
		logrus.Errorf("Error starting gateway: %s", err)
		os.Exit(-1)
	}

	// Create the network listener to attach to the gateway service
	if cfg.NoTLS {
		// just use the regular listener
		logrus.Warn("TLS disabled")
	} else {
		logrus.Infof("Listening requiring TLS")
	}

	// serve websocket connections
	if cfg.WssPort > 0 && !cfg.NoWS {
		lisWS, err = listener.CreateListener(ipAddr, cfg.WssPort, cfg.NoTLS, serverCert, caCert)
		if err == nil {
			go capnpserver.StartGatewayCapnpServer(svc, lisWS, cfg.WssPath)
		}
	}

	// serve TCP connections
	if err == nil {
		lisTcp, err = listener.CreateListener(ipAddr, cfg.TcpPort, cfg.NoTLS, serverCert, caCert)
		if err == nil {
			// this blocks until done
			go capnpserver.StartGatewayCapnpServer(svc, lisTcp, "")
		}
	}
	if err != nil {
		logrus.Fatalf("Gateway startup error: %s", err)
		return
	}

	// serve DNS-SD discovery
	if !cfg.NoDiscovery {
		discoAddr := ipAddr
		if ipAddr == "" {
			discoAddr = listener.GetOutboundIP("").String()
		}
		// addr, _, _ := net.SplitHostPort(lisTcp.Addr().String())
		dnsSrv, err2 := listener.ServeDiscovery(
			serviceName, "hiveot", discoAddr, cfg.TcpPort, cfg.WssPort, cfg.WssPath)
		if err2 == nil {
			defer dnsSrv.Shutdown()
		}
	}
	// wait for the shutdown signal and close down
	listener.WaitForSignal(context.Background())
	_ = lisTcp.Close()
	if lisWS != nil {
		_ = lisWS.Close()
	}
	err = svc.Stop()

	if err == nil {
		logrus.Warningf("%s service has shutdown gracefully", serviceName)
		os.Exit(0)
	} else {
		logrus.Errorf("%s service shutdown with error: %s", serviceName, err)
		os.Exit(-1)
	}
}
