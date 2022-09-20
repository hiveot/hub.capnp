// Package main with the thing store
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hiveot/hub/internal/listener"
	"github.com/hiveot/hub/pkg/directorystore/adapter"

	"github.com/hiveot/hub/pkg/directorystore/thingkvstore"
)

// ServiceName is the name of the store for logging
const ServiceName = "thingstore"

// ThingStorePath is the path to the storage file for the in-memory store.
const ThingStorePath = "config/thingstore.json"

// Use the commandline option -f path/to/store.json for the storage file
func main() {
	thingStorePath := ThingStorePath
	flag.StringVar(&thingStorePath, "f", thingStorePath, "File path of the Thing store.")

	lis := listener.CreateServiceListener(ServiceName)

	store, err := thingkvstore.NewThingKVStoreServer(thingStorePath)
	if err != nil {
		log.Fatalf("Service '%s' failed to start: %s", ServiceName, err)
	}

	adapter.StartDirectoryStoreCapnpAdapter(context.Background(), lis, store)
}
