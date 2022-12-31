package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/service/aclstore"
)

// AuthzService handles client management and authorization for access to Things.
// This implements the IAuthz interface
//
// Authorization uses access control lists with group membership and roles to determine if a client
// is authorized to receive or post a message. This applies to all users of the message bus,
// regardless of how they are authenticated.
type AuthzService struct {
	aclStore *aclstore.AclFileStore
}

// CapClientAuthz returns the capability to verify client authorization
func (authzService *AuthzService) CapClientAuthz(
	ctx context.Context, clientID string) (authz.IClientAuthz, error) {

	verifyAuthz, err := authzService.CapVerifyAuthz(ctx, clientID)
	clientAuthz := NewClientAuthz(clientID, verifyAuthz)
	return clientAuthz, err
}

// CapManageAuthz returns the capability to manage authorization
func (authzService *AuthzService) CapManageAuthz(
	_ context.Context, clientID string) (authz.IManageAuthz, error) {

	manageAuthz := NewManageAuthz(authzService.aclStore, clientID)
	return manageAuthz, nil
}

// CapVerifyAuthz returns the capability to verify authorization
func (authzService *AuthzService) CapVerifyAuthz(
	_ context.Context, clientID string) (authz.IVerifyAuthz, error) {

	verifyAuthz := NewVerifyAuthz(authzService.aclStore, clientID)
	return verifyAuthz, nil
}

// Stop closes the service and release resources
func (authzService *AuthzService) Stop() {
	authzService.aclStore.Close()
}

// Start the ACL store for reading
func (authzService *AuthzService) Start(ctx context.Context) error {
	logrus.Infof("Opening ACL store")
	err := authzService.aclStore.Open(ctx)
	if err != nil {
		return err
	}
	return nil
}

// NewAuthzService creates a new instance of the authorization service.
//
//	aclStore provides the functions to read and write authorization rules
func NewAuthzService(aclStorePath string) *AuthzService {
	aclStore := aclstore.NewAclFileStore(aclStorePath, authz.ServiceName)

	authzService := AuthzService{
		aclStore: aclStore,
	}
	return &authzService
}
