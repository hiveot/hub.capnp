// Package client that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// DeviceCertsCapnpClient is the POGO client to creating device certificates
type DeviceCertsCapnpClient struct {
	capability hubapi.CapDeviceCerts // capnp client for certificates
}

// CreateDeviceCert creates a CA signed certificate for mutual authentication between Hub and IoT devices
func (cl *DeviceCertsCapnpClient) CreateDeviceCert(
	ctx context.Context, deviceID string, pubKeyPEM string, validityDays int) (
	certPEM string, caCertPEM string, err error) {

	// create the method to invoke with the parameters
	createDeviceCertMethod, release := cl.capability.CreateDeviceCert(ctx,
		func(params hubapi.CapDeviceCerts_createDeviceCert_Params) error {
			err2 := params.SetDeviceID(deviceID)
			_ = params.SetPubKeyPEM(pubKeyPEM)
			params.SetValidityDays(int32(validityDays))
			return err2

		})
	defer release()
	// invoke the method and get the result
	resp, err := createDeviceCertMethod.Struct()
	if err == nil {
		certPEM, err = resp.CertPEM()
		caCertPEM, _ = resp.CaCertPEM()
	}
	return certPEM, caCertPEM, err
}

// NewDeviceCertsCapnpClient returns the device certificate client using the capnp protocol
// This is for internal use. The capability has to be obtained using CertsCapnpClient.
func NewDeviceCertsCapnpClient(cap hubapi.CapDeviceCerts) *DeviceCertsCapnpClient {
	cl := &DeviceCertsCapnpClient{capability: cap}
	return cl
}
