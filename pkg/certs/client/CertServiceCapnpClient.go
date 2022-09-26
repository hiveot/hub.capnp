// Package client that wraps the capnp generated client with a POGS API
package client

import (
	"context"
	"net"
	"time"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/internal/caphelp"
)

// CertServiceCapnpClient provides the POGS wrapper around the Capnp API
type CertServiceCapnpClient struct {
	connection *rpc.Conn             // connection to capnp server
	capability hubapi.CapCertService // capnp client
	ctx        context.Context
	ctxCancel  context.CancelFunc
}

// GetDeviceCertCapability returns the capability to create device certificates
//func (cl *CertServiceCapnpClient) GetDeviceCertCapability() (cap hubapi.CapDeviceCert, release capnp.ReleaseFunc, err error) {
//
//	// Get the capability for creating a device certificate for the given device
//	getDeviceCertCap, release := cl.capability.GetDeviceCertCapability(cl.ctx, nil)
//	resp, err := getDeviceCertCap.Struct()
//	capability := resp.Cap()
//	return capability, release, err
//}

// CreateDeviceCert creates a CA signed certificate for mutual authentication between Hub and IoT devices
func (cl *CertServiceCapnpClient) CreateDeviceCert(deviceID string, pubKeyPEM string, validityDays int) (
	certPEM string, caCertPEM string, err error) {

	// First, get the capability for creating a device certificate for the given device
	getCap, release := cl.capability.GetDeviceCertCapability(cl.ctx, nil)
	resp, err := getCap.Struct()
	capability := resp.Cap()

	// create the method to invoke with the parameters
	createDeviceCertMethod, release2 := capability.CreateDeviceCert(cl.ctx,
		func(params hubapi.CapDeviceCert_createDeviceCert_Params) error {
			err2 := params.SetDeviceID(deviceID)
			params.SetPubKeyPEM(pubKeyPEM)
			params.SetValidityDays(int32(validityDays))
			return err2

		})
	// invoke the method and get the result
	resp2, err := createDeviceCertMethod.Struct()
	if err == nil {
		certPEM, err = resp2.CertPEM()
		caCertPEM, _ = resp2.CaCertPEM()
	}
	release2()
	release()

	//method, release := cl.capability.CreateDeviceCert(cl.ctx,
	//	func(params hubapi.ICertService_createDeviceCert_Params) error {
	//		err2 := params.SetDeviceID(deviceID)
	//		params.SetPubKeyPEM(pubKeyPEM)
	//		params.SetValidityDays(int32(validityDays))
	//		return err2
	//	})
	//defer release()
	//resp, err := method.Struct()
	//if err == nil {
	//	certPEM, err = resp.CertPEM()
	//	caCertPEM, _ = resp.CaCertPEM()
	//}
	return certPEM, caCertPEM, err
}

func (cl *CertServiceCapnpClient) CreateServiceCert(serviceID string, pubKeyPEM string, names []string, validityDays int) (
	certPEM string, caCertPEM string, err error) {

	// First, get the capability for creating a service certificate
	getCap, release := cl.capability.GetServiceCertCapability(cl.ctx, nil)
	resp, err := getCap.Struct()
	capability := resp.Cap()

	// Next invoke the method
	method, release := capability.CreateServiceCert(cl.ctx,
		func(params hubapi.CapServiceCert_createServiceCert_Params) error {
			err2 := params.SetServiceID(serviceID)
			params.SetPubKeyPEM(pubKeyPEM)
			if names != nil {
				params.SetNames(caphelp.StringsToCapnp(names))
			}
			params.SetValidityDays(int32(validityDays))
			return err2
		})
	defer release()
	resp2, err := method.Struct()
	if err == nil {
		certPEM, err = resp2.CertPEM()
		caCertPEM, _ = resp2.CaCertPEM()
	}
	return certPEM, caCertPEM, err
}

// CreateUserCert creates a CA signed certificate for mutual authentication by consumers
func (cl *CertServiceCapnpClient) CreateUserCert(clientID string, pubKeyPEM string, validityDays int) (
	certPEM string, caCertPEM string, err error) {

	// First, get the capability for creating a user certificate
	getCap, release := cl.capability.GetUserCertCapability(cl.ctx, nil)
	resp, err := getCap.Struct()
	capability := resp.Cap()

	method, release := capability.CreateUserCert(cl.ctx,
		func(params hubapi.CapUserCert_createUserCert_Params) error {
			err2 := params.SetClientID(clientID)
			params.SetPubKeyPEM(pubKeyPEM)
			params.SetValidityDays(int32(validityDays))
			return err2
		})
	defer release()
	resp2, err := method.Struct()
	if err == nil {
		certPEM, err = resp2.CertPEM()
		caCertPEM, _ = resp2.CaCertPEM()
	}
	return certPEM, caCertPEM, err
}

// VerifyCert verifies is the given certificate is valid
func (cl *CertServiceCapnpClient) VerifyCert(clientID string, certPEM string) (err error) {
	// First, get the capability for verifying a  certificate
	getCap, release := cl.capability.GetVerifyCertCapability(cl.ctx, nil)
	//resp, err := getCap.Struct()
	//capability := resp.Cap()
	// this next line avoids a round trip by using the capability 'future/promise'
	capability := getCap.Cap()

	method, release := capability.VerifyCert(cl.ctx,
		func(params hubapi.CapVerifyCert_verifyCert_Params) error {
			err2 := params.SetClientID(clientID)
			params.SetCertPEM(certPEM)
			return err2
		})
	defer release()
	_, err = method.Struct()
	return err

}

// NewCertServiceCapnpClient returns a capability to create certificates using the capnp protocol
func NewCertServiceCapnpClient(address string, isUDS bool) (*CertServiceCapnpClient, error) {
	network := "tcp"
	if isUDS {
		network = "unix"
	}
	connection, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	transport := rpc.NewStreamTransport(connection)
	rpcConn := rpc.NewConn(transport, nil)
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*60)
	capability := hubapi.CapCertService(rpcConn.Bootstrap(ctx))

	cl := &CertServiceCapnpClient{
		connection: rpcConn,
		capability: capability,
		ctx:        ctx,
		ctxCancel:  ctxCancel,
	}
	return cl, nil
}
