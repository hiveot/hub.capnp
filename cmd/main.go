package main

import (
	"github.com/wostzone/gateway/pkg/gateway"
	"github.com/wostzone/gateway/pkg/lib"
)

func main() {
	gateway.StartGateway("")
	lib.WaitForSignal()
	gateway.StopGateway()
}
