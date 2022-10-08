// Package main with the provisioning service
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/certs"
	certsclient "github.com/hiveot/hub/pkg/certs/capnpclient"
	"github.com/hiveot/hub/pkg/provisioning/capnpserver"
	"github.com/hiveot/hub/pkg/provisioning/service"
)

// ServiceName is the name of the store for logging
const ServiceName = "provisioning"

// FIXME: don't use hard coded socket address for certs service
const CertsSvcAddress = "/tmp/certs.socket"

// Start the provisioning service
// This must be run from a properly setup environment. See GetFolders for details.
func main() {
	var svc *service.ProvisioningService
	var deviceCap certs.IDeviceCerts
	var verifyCap certs.IVerifyCerts
	var certsClient certs.ICerts
	ctx, _ := context.WithCancel(context.Background())

	// Determine the folder layout.
	homeFolder := filepath.Join(filepath.Dir(os.Args[0]), "../..")
	f := folders.GetFolders(homeFolder, false)
	flag.Parse()

	// connect to the certificate service to get its capability for issuing device certificates
	certConn, err := listener.CreateClientConnection(f.Run, certs.ServiceName)
	if err == nil {
		certsClient, err = certsclient.NewCertServiceCapnpClient(ctx, certConn)
	}
	if err == nil {
		deviceCap = certsClient.CapDeviceCerts()
		verifyCap = certsClient.CapVerifyCerts()
	}
	// now we have the capability to create certificates, create the service and start listening for capnp clients
	if err == nil {
		svc = service.NewProvisioningService(deviceCap, verifyCap)
		srvListener := listener.CreateServiceListener(f.Run, ServiceName)
		err = capnpserver.StartProvisioningCapnpServer(context.Background(), srvListener, svc)
	}
	if err != nil {
		log.Fatalf("Service '%s' failed to start: %s", ServiceName, err)
	}
}
