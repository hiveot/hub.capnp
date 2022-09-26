// Package main with the provisioning service
package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"log"
	"path"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/certsclient"
	"github.com/hiveot/hub/internal/folders"
	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/provisioning/service"

	"github.com/hiveot/hub/pkg/provisioning/service/oobprovserver"
)

// ServiceName is the name of the store for logging
const ServiceName = "provisioning"

// Start the provisioning service
func main() {
	var caKey *ecdsa.PrivateKey
	var svc *oobprovserver.OobProvServer

	certFolder := folders.GetFolders("").Certs
	flag.StringVar(&certFolder, "certs", certFolder, "Certificate folder.")

	lis := listener.CreateServiceListener(ServiceName)
	caCertPath := path.Join(certFolder, hubapi.DefaultCaCertFile)
	caKeyPath := path.Join(certFolder, hubapi.DefaultCaKeyFile)
	caCert, err := certsclient.LoadX509CertFromPEM(caCertPath)
	if err == nil {
		caKey, err = certsclient.LoadKeysFromPEM(caKeyPath)
	}
	if err == nil {
		svc, err = oobprovserver.NewOobProvServer(caCert, caKey)
	}
	if err == nil {
		err = service.StartProvisioningCapnpAdapter(context.Background(), lis, svc)
	}
	if err != nil {
		log.Fatalf("Service '%s' failed to start: %s", ServiceName, err)
	}
}
