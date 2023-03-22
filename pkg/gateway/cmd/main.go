package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"github.com/sirupsen/logrus"
	"net"
	"os"

	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/certs"
	"github.com/hiveot/hub/pkg/certs/capnpclient"
	"github.com/hiveot/hub/pkg/certs/service/selfsigned"
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
		ipAddr = listener.GetOutboundIP("").String()
	}

	// certificates are needed for TLS connections to the capnp server.
	// on each restart a new set of keys is used and a new certificate is requested.
	keys := certsclient.CreateECDSAKeys()
	serverCert, caCert, err := RenewServiceCerts(serviceName, ipAddr, keys, f.Run)
	if err != nil {
		logrus.Panicf("certs service not reachable when starting the gateway: %s", err)
	}

	// need certs but not authn
	if err == nil {
		//resolverPath := path.Join(f.Run, resolver.ServiceName+".socket")
		resolverPath := resolver.DefaultResolverPath
		// get the authn service from the resolver when needed
		svc = service.NewGatewayService(resolverPath, nil)
	}
	err = svc.Start()
	if err != nil {
		logrus.Errorf("Error starting gateway: %s", err)
		os.Exit(-1)
	}

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
	} else if !cfg.NoDiscovery {
		// DNS-SD discovery
		addr, _, _ := net.SplitHostPort(lisTcp.Addr().String())
		dnsSrv, err2 := listener.ServeDiscovery(
			serviceName, "hiveot", addr, cfg.TcpPort, cfg.WssPort, cfg.WssPath)
		if err2 == nil {
			defer dnsSrv.Shutdown()
		}
	}

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

// RenewServiceCerts obtains a new service certificate from the certs service
// This returns a service certificate signed by the CA, and the certificate of
// the CA that signed the service cert.
// This panics if the certs service is not reachable
//
//	serviceID is the instance ID of the service used as the CN on the certificate
//	ipAddr ip address the service is listening on or "" for outbound IP
//	pubKeyPEM is the public key for the certificate
//	socketFolder is the location of the certs service socket
func RenewServiceCerts(serviceID string, ipAddr string, keys *ecdsa.PrivateKey, socketFolder string) (
	svcCert *tls.Certificate, caCert *x509.Certificate, err error) {
	var capServiceCert certs.IServiceCerts
	if ipAddr == "" {
		ip := listener.GetOutboundIP("")
		ipAddr = ip.String()
	}
	names := []string{"127.0.0.1", ipAddr}
	pubKeyPEM, err := certsclient.PublicKeyToPEM(&keys.PublicKey)
	if err != nil {
		logrus.Errorf("invalid public key: %s", err)
		return nil, nil, err
	}

	ctx := context.Background()
	csConn, err := hubclient.ConnectToUDS(certs.ServiceName, socketFolder)
	if err != nil {
		logrus.Errorf("unable to connect to certs service: %s. Workaround with local instance", err)
		// FIXME: workaround or panic?
		capServiceCert = selfsigned.NewServiceCertsService(caCert, nil)
		return nil, nil, err
	} else {
		cs := capnpclient.NewCertServiceCapnpClient(csConn)
		capServiceCert, err = cs.CapServiceCerts(ctx, hubapi.AuthTypeService)
		_ = err
	}
	svcPEM, caPEM, err := capServiceCert.CreateServiceCert(ctx, serviceID, pubKeyPEM, names, 0)
	if err != nil {
		logrus.Errorf("unable to create a service certificate: %s", err)
		return nil, nil, err
	}
	caCert, _ = certsclient.X509CertFromPEM(caPEM)
	privKeyPEM, _ := certsclient.PrivateKeyToPEM(keys)
	newSvcCert, err := tls.X509KeyPair([]byte(svcPEM), []byte(privKeyPEM))
	return &newSvcCert, caCert, err
}
