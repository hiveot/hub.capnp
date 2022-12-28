package service

import (
	"context"
	"crypto/ecdsa"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.go/pkg/signing"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/config"
	"github.com/hiveot/hub/pkg/authn/service/jwtauthn"
	"github.com/hiveot/hub/pkg/authn/service/unpwstore"
)

// AuthnService provides the capabilities to manage and use authentication services
// This implements the IAuthnService interface
type AuthnService struct {
	config config.AuthnConfig
	// key used for signing of JWT tokens
	signingKey *ecdsa.PrivateKey
	// password storage
	pwStore unpwstore.IUnpwStore
}

// CapUserAuthn creates a new user authentication session
func (svc *AuthnService) CapUserAuthn(ctx context.Context, clientID string) (authn.IUserAuthn, error) {
	_ = ctx
	jwtAuthn := jwtauthn.NewJWTAuthn(
		svc.signingKey, svc.config.AccessTokenValiditySec, svc.config.RefreshTokenValiditySec)
	capUserAuthn := NewUserAuthn(clientID, jwtAuthn, svc.pwStore)
	return capUserAuthn, nil
}

// CapManageAuthn creates a new manage authentication session
func (svc *AuthnService) CapManageAuthn(ctx context.Context, clientID string) (authn.IManageAuthn, error) {
	_ = ctx
	capManageAuthn := NewManageAuthn(svc.pwStore)
	return capManageAuthn, nil
}

func (svc *AuthnService) Start(ctx context.Context) error {
	logrus.Infof("starting authn service using '%s' for password store", svc.config.PasswordFile)
	return svc.pwStore.Open(ctx)
}
func (svc *AuthnService) Stop() error {
	logrus.Info("stopping service")
	svc.pwStore.Close()
	return nil
}

// NewAuthnService creates new instance of the service.
// Call Connect before using the service.
func NewAuthnService(cfg config.AuthnConfig) *AuthnService {
	signingKey := signing.CreateECDSAKeys()
	pwStore := unpwstore.NewPasswordFileStore(cfg.PasswordFile)
	svc := &AuthnService{
		config:     cfg,
		pwStore:    pwStore,
		signingKey: signingKey,
	}
	return svc
}
