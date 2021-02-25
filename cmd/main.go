package main

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/gateway"
	"github.com/wostzone/gateway/pkg/lib"
)

func main() {
	err := gateway.StartGateway("")
	if err != nil {
		logrus.Fatalf("main: Failed starting gateway: %s", err)
	}
	lib.WaitForSignal()
	gateway.StopGateway()
}
