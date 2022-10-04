// Package capnpserver with the capnproto server for the CapCerts API
package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/certs"
)

// UserCertsCapnpServer provides the capnpr RPC server for user certificates
// See hub.capnp/go/hubapi/Cert.capnp.go for the interface
type UserCertsCapnpServer struct {
	srv certs.IUserCerts
}

// CreateUserCert provides the capnp RPC handler for creating user certificates
func (capsrv *UserCertsCapnpServer) CreateUserCert(
	ctx context.Context, call hubapi.CapUserCerts_createUserCert) error {

	clientID, _ := call.Args().ClientID()
	pubKeyPEM, _ := call.Args().PubKeyPEM()
	validityDays := call.Args().ValidityDays()
	if validityDays == 0 {
		validityDays = hubapi.DefaultClientCertValidityDays
	}
	certPEM, caCertPEM, err := capsrv.srv.CreateUserCert(ctx, clientID, pubKeyPEM, int(validityDays))
	if err == nil {
		//logrus.Infof("Created client cert for %s", clientID)
		res, err2 := call.AllocResults()
		if err2 == nil {
			err2 = res.SetCertPEM(certPEM)
			_ = res.SetCaCertPEM(caCertPEM)
		}
		err = err2
	}
	return err
}
