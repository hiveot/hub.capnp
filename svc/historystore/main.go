// Package main with the history store
package main

import (
	"log"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/wostzone/wost.grpc/go/svc"
	"svc/historystore/mongohs"
	"svc/internal/listener"
)

const ServiceName = "historystore"

// Start the history store service using gRPC
func main() {
	lis := listener.CreateServiceListener(ServiceName)

	s := grpc.NewServer()
	service := &mongohs.MongoHistoryStoreServer{}
	svc.RegisterHistoryStoreServer(s, service)

	// exit the service when signal is received and close the listener
	listener.ExitOnSignal(lis, func() {
		logrus.Infof("Shutting down '%s'", ServiceName)
	})

	// Start listening
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Service '%s; exited: %v", ServiceName, err)
	}
}
