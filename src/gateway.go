// Package src with gateway source
package src

import "github.com/wostzone/gateway/src/servicebus"

const hostname = "localhost:9678"

// StartGateway launches the gateway
func StartGateway() {
	println("Starting the WoST Gateway (once it is implemented)")
	// read config
	// determine service bus to use
	// launch internal service bus
	servicebus.StartServiceBus(hostname, nil)
	// launch plugins

}
