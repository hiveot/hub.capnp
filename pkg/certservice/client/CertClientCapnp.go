// Package client with capnp client for cert service
package client

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
)

// CertClientCapnp is a pogo client for the capnp client of the certificate service
// app -> CertClientCapnp -> capnp.CertService (client) -> capnp.CertServer -> CertServiceCapnpAdapter -> CertService
type CertClientCapnp struct {
	capability *hubapi.CertService
	ctx        context.Context
}

// CreateClientCert creates a CA signed certificate for mutual authentication by consumers
func (cl *CertClientCapnp) CreateClientCert(
	clientID string, pubKeyPEM string, validityDays int) (certPEM string, caCertPEM string, err error) {

	method, release := cl.capability.CreateClientCert(cl.ctx,
		func(params hubapi.CertService_createClientCert_Params) error {
			err2 := params.SetClientID(clientID)
			params.SetPubKeyPEM(pubKeyPEM)
			params.SetValidityDays(int32(validityDays))
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err != nil {
		return "", "", err
	}
	certPEM, err = resp.CertPEM()
	caCertPEM, _ = resp.CaCertPEM()
	return certPEM, caCertPEM, err
}

// CreateDeviceCert creates a CA signed certificate for mutual authentication between Hub and IoT devices
func (cl *CertClientCapnp) CreateDeviceCert(
	deviceID string, pubKeyPEM string, validityDays int) (
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
	if err != nil {
		return "", "", err
	}
	certPEM, err = resp.CertPEM()
	caCertPEM, _ = resp.CaCertPEM()
	return certPEM, caCertPEM, err
}

func (cl *CertClientCapnp) CreateServiceCert(
	serviceID string, pubKeyPEM string, names []string, validityDays int) (
	certPEM string, caCertPEM string, err error) {
	method, release := cl.capability.CreateServiceCert(cl.ctx,
		func(params hubapi.CertService_createServiceCert_Params) error {
			err2 := params.SetServiceID(serviceID)
			params.SetPubKeyPEM(pubKeyPEM)
			if names != nil {
				namesCap, _ := params.Names()
				for i := 0; i < len(names); i++ {
					namesCap.Set(i, names[i])
				}
			}
			params.SetValidityDays(int32(validityDays))
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err != nil {
		return "", "", err
	}
	certPEM, err = resp.CertPEM()
	caCertPEM, _ = resp.CaCertPEM()
	return certPEM, caCertPEM, err

}

// NewCertClientCapnp returns a certificate service client wrapper for the capnp protocol
func NewCertClientCapnp(capability *hubapi.CertService) *CertClientCapnp {
	return &CertClientCapnp{
		capability: capability,
	}
}
