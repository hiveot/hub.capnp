// Package main with the provisioning service
package main

import (
	"context"
	"net"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/certs"
	certsclient "github.com/hiveot/hub/pkg/certs/capnpclient"
	"github.com/hiveot/hub/pkg/provisioning"
	"github.com/hiveot/hub/pkg/provisioning/capnpserver"
	"github.com/hiveot/hub/pkg/provisioning/service"
)

// Start the provisioning service
// - dependent on certs service
func main() {
	var svc *service.ProvisioningService
	var deviceCap certs.IDeviceCerts
	var verifyCap certs.IVerifyCerts
	var certsClient certs.ICerts
	ctx := context.Background()

	// Determine the folder layout and handle commandline options
	f := svcconfig.LoadServiceConfig(provisioning.ServiceName, false, nil)

	// connect to the certificate service to get its capability for issuing device certificates
	certConn, err := listener.CreateLocalClientConnection(certs.ServiceName, f.Run)
	if err == nil {
		certsClient = certsclient.NewCertServiceCapnpClient(certConn)
		// the provisioning service requires certificate capabilities
		deviceCap = certsClient.CapDeviceCerts(ctx)
		verifyCap = certsClient.CapVerifyCerts(ctx)
		svc = service.NewProvisioningService(deviceCap, verifyCap)
	}
	// now we have the capability to create certificates, start the service and start listening for capnp clients
	listener.RunService(provisioning.ServiceName, f.SocketPath,
		func(ctx context.Context, lis net.Listener) error {
			// startup
			err := svc.Start(ctx)
			if err == nil {
				err = capnpserver.StartProvisioningCapnpServer(lis, svc)
			}
			return err
		}, func() error {
			// shutdown
			err := svc.Stop()
			return err
		})
}
