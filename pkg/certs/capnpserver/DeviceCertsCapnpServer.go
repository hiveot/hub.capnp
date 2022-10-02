// Package capnpserver with the capnproto server for the CapCerts API
package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/certs"
)

// DeviceCertsCapnpServer provides the capnp RPC server for device certificates
// TODO: option to restrict capability to a specific deviceID
type DeviceCertsCapnpServer struct {
	srv certs.IDeviceCerts
	// TODO: option to restrict to a specific device
	//deviceID string
}

// CreateDeviceCert provides the capnp RPC handler for creating IoT device certificates
func (capsrv *DeviceCertsCapnpServer) CreateDeviceCert(
	ctx context.Context, call hubapi.CapDeviceCerts_createDeviceCert) error {

	deviceID, _ := call.Args().DeviceID()

	pubKeyPEM, _ := call.Args().PubKeyPEM()
	validityDays := call.Args().ValidityDays()
	if validityDays == 0 {
		validityDays = hubapi.DefaultClientCertValidityDays
	}
	certPEM, caCertPEM, err := capsrv.srv.CreateDeviceCert(ctx, deviceID, pubKeyPEM, int(validityDays))
	if err == nil {
		//logrus.Infof("Created device cert for %s", deviceID)
		res, err2 := call.AllocResults()
		if err2 == nil {
			err2 = res.SetCertPEM(certPEM)
			_ = res.SetCaCertPEM(caCertPEM)
		}
		err = err2
	}
	return err
}

// NewDeviceCertsCapnpServer creates a new instance of device certificate capnp server
// For internal use to serve the capnp RPC request for device certificate capability. A new instance
// is created for each client that receives this capability.
func NewDeviceCertsCapnpServer(srv certs.IDeviceCerts) *DeviceCertsCapnpServer {
	capsrv := &DeviceCertsCapnpServer{srv: srv}
	return capsrv
}
