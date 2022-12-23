package captest

import (
	"context"
	"fmt"
)

// Method1ServiceCapnpServer is the capability to run Method1
// This implements the CapMethod1_Server interface
type Method1ServiceCapnpServer struct {
	clientID   string
	clientType string
}

func (m1 *Method1ServiceCapnpServer) Method1(_ context.Context, params CapMethod1Service_method1) error {
	args := params.Args()
	_ = args
	resp, _ := params.AllocResults()
	err := resp.SetForYou(fmt.Sprintf("Hello '%s', capnproto is great!", m1.clientID))
	return err
}

func NewMethod1ServiceCapnpServer(clientID, clientType string) *Method1ServiceCapnpServer {
	srv := &Method1ServiceCapnpServer{
		clientID:   clientID,
		clientType: clientType,
	}
	return srv
}
