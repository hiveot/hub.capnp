package selfsigned

import (
	"github.com/wostzone/wost.grpc/go/svc"
)

// SelfSignedServer implements the svc.CertServiceServer interface
// This service creates certificates for use by services, devices (via idprov) and admin users.
type SelfSignedServer struct {
	svc.UnimplementedCertServiceServer
}
