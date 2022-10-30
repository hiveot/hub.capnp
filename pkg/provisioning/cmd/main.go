// Package main with the provisioning service
package main

import (
	"context"
	"log"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/certs"
	certsclient "github.com/hiveot/hub/pkg/certs/capnpclient"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/provisioning"
	"github.com/hiveot/hub/pkg/provisioning/capnpserver"
	"github.com/hiveot/hub/pkg/provisioning/service"
)

// Start the provisioning service
// This must be run from a properly setup environment. See GetFolders for details.
func main() {
	var svc *service.ProvisioningService
	var deviceCap certs.IDeviceCerts
	var verifyCap certs.IVerifyCerts
	var certsClient certs.ICerts
	ctx, _ := context.WithCancel(context.Background())

	// Determine the folder layout and handle commandline options
	f := svcconfig.LoadServiceConfig(launcher.ServiceName, false, nil)

	// connect to the certificate service to get its capability for issuing device certificates
	certConn, err := listener.CreateClientConnection(f.Run, certs.ServiceName)
	if err == nil {
		certsClient, err = certsclient.NewCertServiceCapnpClient(ctx, certConn)
	}
	// the provisioning service requires certificate capabilities
	if err == nil {
		deviceCap = certsClient.CapDeviceCerts()
		verifyCap = certsClient.CapVerifyCerts()
	}
	// now we have the capability to create certificates, create the service and start listening for capnp clients
	if err == nil {
		svc = service.NewProvisioningService(ctx, deviceCap, verifyCap)
		srvListener := listener.CreateServiceListener(f.Run, provisioning.ServiceName)
		err = capnpserver.StartProvisioningCapnpServer(context.Background(), srvListener, svc)
	}
	if err != nil {
		log.Fatalf("Service '%s' failed to start: %s", provisioning.ServiceName, err)
	}
}
