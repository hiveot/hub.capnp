// Package gateway with gateway source
package gateway

const hostname = "localhost:9678"

// StartGateway launches the gateway
func StartGateway() {
	println("Starting the WoST Gateway (once it is implemented)")
	// read config
	// determine service bus to use
	// launch internal service bus
	StartSimbuServer(hostname)
	// launch plugins

}
