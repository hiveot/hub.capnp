package captest

import (
	"context"
	"fmt"
)

// Method1Service is the capability to run Method1
// This implements the CapMethod1_Server interface
type Method1Service struct {
	clientID   string
	clientType string
}

func (m1 *Method1Service) Method1(_ context.Context, params CapMethod1Service_method1) error {
	args := params.Args()
	_ = args
	resp, _ := params.AllocResults()
	err := resp.SetForYou(fmt.Sprintf("Hello '%s', capnproto is great!", m1.clientID))
	return err
}
