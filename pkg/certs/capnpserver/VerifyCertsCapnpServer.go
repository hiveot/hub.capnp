// Package capnpserver with the capnproto server for the CapCerts API
package capnpserver

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/certs"
)

// VerifyCertsCapnpServer provides the capnpr RPC server for certificate verification
type VerifyCertsCapnpServer struct {
	srv certs.IVerifyCerts
}

func (capsrv *VerifyCertsCapnpServer) VerifyCert(
	ctx context.Context, call hubapi.CapVerifyCerts_verifyCert) error {

	clientID, _ := call.Args().ClientID()
	certPEM, _ := call.Args().CertPEM()
	err := capsrv.srv.VerifyCert(ctx, clientID, certPEM)
	return err
}
