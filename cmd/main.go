package main

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/gateway"
)

func main() {
	err := gateway.StartGateway("", true)
	if err != nil {
		logrus.Fatalf("main: Failed starting gateway: %s", err)
	}
	gateway.WaitForSignal()
	gateway.StopGateway()
}
