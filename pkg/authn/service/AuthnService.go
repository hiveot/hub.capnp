package service

import (
	"context"
	"crypto/ecdsa"

	"github.com/hiveot/hub.go/pkg/signing"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/config"
	"github.com/hiveot/hub/pkg/authn/service/jwtauthn"
	"github.com/hiveot/hub/pkg/authn/service/unpwstore"
)

// AuthnService provides the capabilities to manage and use authentication services
// This implements the IAuthn interface
type AuthnService struct {
	config config.AuthnConfig
	// key used for signing of JWT tokens
	signingKey *ecdsa.PrivateKey
	// password storage
	pwStore unpwstore.IUnpwStore
}

func (svc *AuthnService) CapUserAuthn(ctx context.Context, loginID string) authn.IUserAuthn {
	_ = ctx
	jwtAuthn := jwtauthn.NewJWTAuthn(
		svc.signingKey, svc.config.AccessTokenValiditySec, svc.config.RefreshTokenValiditySec)
	capUserAuthn := NewUserAuthn(loginID, jwtAuthn, svc.pwStore)
	return capUserAuthn
}

func (svc *AuthnService) CapManageAuthn(ctx context.Context) authn.IManageAuthn {
	_ = ctx
	capManageAuthn := NewManageAuthn(svc.pwStore)
	return capManageAuthn
}

func (svc *AuthnService) Start(ctx context.Context) error {
	return svc.pwStore.Open(ctx)
}
func (svc *AuthnService) Stop(ctx context.Context) error {
	svc.pwStore.Close()
	return nil
}

// NewAuthnService creates new instance of the service.
// Call Start before using the service.
func NewAuthnService(ctx context.Context, cfg config.AuthnConfig) *AuthnService {
	signingKey := signing.CreateECDSAKeys()
	pwStore := unpwstore.NewPasswordFileStore(cfg.PasswordFile)
	svc := &AuthnService{
		config:     cfg,
		pwStore:    pwStore,
		signingKey: signingKey,
	}
	return svc
}
