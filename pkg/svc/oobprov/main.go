// Package main with the provisioning service
package main

import (
	"flag"
	"log"
	"path"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/wostzone/hub/internal/folders"
	"github.com/wostzone/hub/internal/listener"

	"github.com/wostzone/hub/pkg/svc/oobprov/oobprovserver"
	"github.com/wostzone/wost.grpc/go/svc"
)

// ServiceName is the name of the store for logging
const ServiceName = "provisioning"

// Start the gRPC provisioning service
func main() {
	certFolder := folders.GetFolders("").Certs
	flag.StringVar(&certFolder, "certs", certFolder, "Certificate folder.")

	lis := listener.CreateServiceListener(ServiceName)
	caCertPath := path.Join(certFolder, service.DefaultCaCertFile)
	caKeyPath := path.Join(certFolder, service.DefaultCaKeyFile)

	service, err := oobprovserver.NewOobProvServer(caCertPath, caKeyPath)
	if err != nil {
		log.Fatalf("Service '%s' failed to start: %s", ServiceName, err)
	}

	s := grpc.NewServer()
	svc.RegisterProvisioningServer(s, service)

	// exit the service when signal is received and close the listener
	listener.ExitOnSignal(lis, func() {
		logrus.Infof("Shutting down '%s'", ServiceName)
	})

	// Start listening
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Service '%s; exited: %v", ServiceName, err)
	}
}
