package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/lib/listener"
	"github.com/hiveot/hub/lib/svcconfig"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/hiveot/hub/pkg/resolver/capnpserver"
	"github.com/hiveot/hub/pkg/resolver/service"
)

// main launches the resolver service
func main() {
	//resolverSocketPath := resolver.DefaultResolverPath

	f := svcconfig.LoadServiceConfig(resolver.ServiceName, false, nil)
	svc := service.NewResolverService(f.Run)

	// the resolver uses unix sockets to listen for incoming connections
	listener.RunService(resolver.ServiceName, resolver.DefaultResolverPath, //f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			if err == nil {
				capnpserver.StartResolverServiceCapnpServer(svc, lis, svc.HandleUnknownMethod)
			}
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})
}

// RenewServiceCerts obtains a new service certificate from the certs service
// This returns a service certificate signed by the CA, and the certificate of
// the CA that signed the service cert.
// This panics if the certs service is not reachable
//
//	serviceID is the instance ID of the service used as the CN on the certificate
//	pubKeyPEM is the public key for the certificate
//	socketFolder is the location of the certs service socket
//func RenewServiceCerts(serviceID string, keys *ecdsa.PrivateKey, socketFolder string) (
//	svcCert *tls.Certificate, caCert *x509.Certificate, err error) {
//
//	ctx := context.Background()
//	csConn, err := listener.CreateLocalClientConnection(certs.ServiceName, socketFolder)
//	if err != nil {
//		return nil, nil, err
//	}
//	cs := capnpclient.NewCertServiceCapnpClient(csConn)
//	capServiceCert := cs.CapServiceCerts(ctx, resolver.ServiceName)
//	ipAddr := hubnet.GetOutboundIP("")
//	names := []string{"127.0.0.1", ipAddr.String()}
//	pubKeyPEM, _ := certsclient.PublicKeyToPEM(keys.PublicKey)
//	svcPEM, caPEM, err := capServiceCert.CreateServiceCert(ctx, serviceID, pubKeyPEM, names, 0)
//	if err != nil {
//		return nil, nil, err
//	}
//	caCert, _ = certsclient.X509CertFromPEM(caPEM)
//	privKeyPEM, _ := certsclient.PrivateKeyToPEM(keys)
//	newSvcCert, err := tls.X509KeyPair([]byte(svcPEM), []byte(privKeyPEM))
//	return &newSvcCert, caCert, err
//}
