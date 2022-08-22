// Package main with the thing store
package main

import (
	"flag"
	"log"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/wostzone/wost.grpc/go/svc"
	"svc/internal/listener"
	"svc/thingstore/thingkvstore"
)

// ServiceName is the name of the store for logging
const ServiceName = "thingstore"

// ThingStorePath is the path to the storage file for the in-memory store.
const ThingStorePath = "config/thingstore.json"

// Start the gRPC history in-memory store service
// Use the commandline option -f path/to/store.json for the storage file
func main() {
	thingStorePath := ThingStorePath
	flag.StringVar(&thingStorePath, "f", thingStorePath, "File path of the Thing store.")

	lis := listener.CreateServiceListener(ServiceName)

	service, err := thingkvstore.NewThingKVStoreServer(thingStorePath)
	if err != nil {
		log.Fatalf("Service '%s' failed to start: %s", ServiceName, err)
	}

	s := grpc.NewServer()
	svc.RegisterThingStoreServer(s, service)

	// exit the service when signal is received and close the listener
	listener.ExitOnSignal(lis, func() {
		logrus.Infof("Shutting down '%s'", ServiceName)
	})

	// Start listening
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Service '%s; exited: %v", ServiceName, err)
	}
}
