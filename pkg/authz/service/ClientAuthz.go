package service

import (
	"context"

	"github.com/hiveot/hub/pkg/authz"
)

// ClientAuthz determines client authorization for access to Things
type ClientAuthz struct {
	clientID    string
	verifyAuthz authz.IVerifyAuthz
}

// GetPermissions returns a list of permissions a client has for a Thing
func (clauthz *ClientAuthz) GetPermissions(
	ctx context.Context, thingID string) (permissions []string, err error) {

	return clauthz.verifyAuthz.GetPermissions(ctx, clauthz.clientID, thingID)
}

// NewClientAuthz creates an instance handler for verifying client specific authorization.
// This is the same as VerifyAuthz but restricted to a specific client
func NewClientAuthz(clientID string, verifyAuthz authz.IVerifyAuthz) *ClientAuthz {
	clauthz := ClientAuthz{
		clientID:    clientID,
		verifyAuthz: verifyAuthz,
	}
	return &clauthz
}
