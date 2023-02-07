package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/lib/certsclient"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
	capnpclient2 "github.com/hiveot/hub/pkg/authn/capnpclient"
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
	var userAuthn authn.IUserAuthn
	var lisTcp net.Listener
	var lisWS net.Listener
	ctx := context.Background()

	f, _, _ := svcconfig.SetupFolderConfig(serviceName)
	cfg := config.NewGatewayConfig()
	_ = f.LoadConfig(&cfg)

	// certificates are needed for TLS connections to the capnp server.
	// on each restart a new set of keys is used and a new certificate is requested.
	keys := certsclient.CreateECDSAKeys()
	serverCert, caCert, err := RenewServiceCerts(serviceName, keys, f.Run)
	if err != nil {
		logrus.Panicf("certs service not reachable when starting the gateway: %s", err)
	}

	// The authn service is used to authenticate logins from users.
	// without authn it still functions with certificates
	//authnConn, err := listener.CreateLocalClientConnection(authn.ServiceName, f.Run)
	// conn, err := hubclient.ConnectToHub("", "", nil, nil)
	fullURL := "unix://" + hubapi.DefaultResolverAddress
	conn, err := hubclient.CreateClientConnection(fullURL, nil, nil)

	if err == nil {
		authnService := capnpclient2.NewAuthnCapnpClient(context.Background(), conn)
		defer authnService.Release()
		userAuthn, err = authnService.CapUserAuthn(ctx, serviceName)
	}
	// need certs but not authn
	if err == nil {
		//resolverPath := path.Join(f.Run, resolver.ServiceName+".socket")
		resolverPath := resolver.DefaultResolverPath
		svc = service.NewGatewayService(resolverPath, userAuthn)
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

	// optionally serve websockets
	if !cfg.NoWS {
		lisWS, err = listener.CreateListener(cfg.WSAddress, cfg.NoTLS, serverCert, caCert)
		if err == nil {
			// the WS runs in the background
			parts := strings.Split(cfg.WSAddress, "/")
			wsPath := "/" + parts[len(parts)-1]
			go capnpserver.StartGatewayCapnpServer(svc, lisWS, wsPath)
		}
	}

	// always listen on tcp
	if err == nil {
		lisTcp, err = listener.CreateListener(cfg.Address, cfg.NoTLS, serverCert, caCert)
		if err == nil {
			err = capnpserver.StartGatewayCapnpServer(svc, lisTcp, "")
		}
	}
	if err != nil {
		logrus.Fatalf("Gateway startup error: %s", err)
	}

	_ = listener.ExitOnSignal(context.Background(), func() {
		_ = lisTcp.Close()
		if lisWS != nil {
			_ = lisWS.Close()
		}
		err = svc.Stop()
	})

	if errors.Is(err, net.ErrClosed) {
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
//	pubKeyPEM is the public key for the certificate
//	socketFolder is the location of the certs service socket
func RenewServiceCerts(serviceID string, keys *ecdsa.PrivateKey, socketFolder string) (
	svcCert *tls.Certificate, caCert *x509.Certificate, err error) {
	var capServiceCert certs.IServiceCerts

	ipAddr := listener.GetOutboundIP("")
	names := []string{"127.0.0.1", ipAddr.String()}
	pubKeyPEM, err := certsclient.PublicKeyToPEM(&keys.PublicKey)
	if err != nil {
		logrus.Errorf("invalid public key: %s", err)
		return nil, nil, err
	}

	ctx := context.Background()
	csConn, err := hubclient.ConnectToService(certs.ServiceName, socketFolder)
	if err != nil {
		logrus.Errorf("unable to connect to certs service: %s. Workaround with local instance", err)
		// FIXME: workaround or panic?
		capServiceCert = selfsigned.NewServiceCertsService(caCert, nil)
		return nil, nil, err
	} else {
		cs := capnpclient.NewCertServiceCapnpClient(csConn)
		capServiceCert, err = cs.CapServiceCerts(ctx, hubapi.ClientTypeService)
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
