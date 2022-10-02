// Package capnpserver with the capnproto server for the CapCerts API
package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/certs"
)

// ServiceCertsCapnpServer provides the capnpr RPC server service certificates
// See hub.capnp/go/hubapi/Cert.capnp.go for the interface
type ServiceCertsCapnpServer struct {
	srv certs.IServiceCerts
}

func (capsrv *ServiceCertsCapnpServer) CreateServiceCert(
	ctx context.Context, call hubapi.CapServiceCerts_createServiceCert) error {
	clientID, _ := call.Args().ServiceID()
	pubKeyPEM, _ := call.Args().PubKeyPEM()
	namesList, _ := call.Args().Names()
	validityDays := call.Args().ValidityDays()
	if validityDays == 0 {
		validityDays = hubapi.DefaultServiceCertValidityDays
	}
	names := []string{}
	for i := 0; i < namesList.Len(); i++ {
		name, _ := namesList.At(i)
		names = append(names, name)
	}
	certPEM, caCertPEM, err := capsrv.srv.CreateServiceCert(ctx, clientID, pubKeyPEM, names, int(validityDays))
	if err == nil {
		//logrus.Infof("Created device cert for %s", clientID)
		res, err2 := call.AllocResults()
		if err2 == nil {
			err2 = res.SetCertPEM(certPEM)
			_ = res.SetCaCertPEM(caCertPEM)
		}
		err = err2
	}
	return err
}

// NewServiceCertsCapnpServer creates a new instance of service certificate capnp server
// For internal use to serve the capnp request for device certificate capability. A new instance
// is created for each client that receives this capability.
func NewServiceCertsCapnpServer(srv certs.IServiceCerts) *ServiceCertsCapnpServer {
	capsrv := &ServiceCertsCapnpServer{srv: srv}
	return capsrv
}
