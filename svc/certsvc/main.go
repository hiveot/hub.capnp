package main

import (
	"log"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/wostzone/wost.grpc/go/svc"
	"svc/certsvc/selfsigned"
	"svc/internal/listener"
)

const ServiceName = "certsvc"

// Start the service using gRPC
func main() {
	// handle commandline to create a listener
	lis := listener.CreateServiceListener(ServiceName)

	s := grpc.NewServer()
	service := &selfsigned.SelfSignedServer{}
	svc.RegisterCertServiceServer(s, service)

	// exit the service when signal is received and close the listener
	listener.ExitOnSignal(lis, func() {
		logrus.Infof("Shutting down '%s'", ServiceName)
	})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Service '%s; exited: %v", ServiceName, err)
	}
}
