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
	connection *rpc.Conn          // connection to capnp server
	capability hubapi.CertService // capnp client
	ctx        context.Context
	ctxCancel  context.CancelFunc
}

// CreateDeviceCert creates a CA signed certificate for mutual authentication between Hub and IoT devices
func (cl *CertServiceCapnpClient) CreateDeviceCert(deviceID string, pubKeyPEM string, validityDays int) (
	certPEM string, caCertPEM string, err error) {

	method, release := cl.capability.CreateDeviceCert(cl.ctx,
		func(params hubapi.CertService_createDeviceCert_Params) error {
			err2 := params.SetDeviceID(deviceID)
			params.SetPubKeyPEM(pubKeyPEM)
			params.SetValidityDays(int32(validityDays))
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		certPEM, err = resp.CertPEM()
		caCertPEM, _ = resp.CaCertPEM()
	}
	return certPEM, caCertPEM, err
}

func (cl *CertServiceCapnpClient) CreateServiceCert(serviceID string, pubKeyPEM string, names []string, validityDays int) (
	certPEM string, caCertPEM string, err error) {

	method, release := cl.capability.CreateServiceCert(cl.ctx,
		func(params hubapi.CertService_createServiceCert_Params) error {
			err2 := params.SetServiceID(serviceID)
			params.SetPubKeyPEM(pubKeyPEM)
			if names != nil {
				params.SetNames(caphelp.StringsToCapnp(names))
			}
			params.SetValidityDays(int32(validityDays))
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		certPEM, err = resp.CertPEM()
		caCertPEM, _ = resp.CaCertPEM()
	}
	return certPEM, caCertPEM, err
}

// CreateUserCert creates a CA signed certificate for mutual authentication by consumers
func (cl *CertServiceCapnpClient) CreateUserCert(clientID string, pubKeyPEM string, validityDays int) (
	certPEM string, caCertPEM string, err error) {

	method, release := cl.capability.CreateUserCert(cl.ctx,
		func(params hubapi.CertService_createUserCert_Params) error {
			err2 := params.SetClientID(clientID)
			params.SetPubKeyPEM(pubKeyPEM)
			params.SetValidityDays(int32(validityDays))
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		certPEM, err = resp.CertPEM()
		caCertPEM, _ = resp.CaCertPEM()
	}
	return certPEM, caCertPEM, err
}

// NewCertServiceCapnpClient returns a certificate service client wrapper for the capnp protocol
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
	capability := hubapi.CertService(rpcConn.Bootstrap(ctx))

	cl := &CertServiceCapnpClient{
		connection: rpcConn,
		capability: capability,
		ctx:        ctx,
		ctxCancel:  ctxCancel,
	}
	return cl, nil
}
