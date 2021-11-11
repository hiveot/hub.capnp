package dirserver

import (
	"github.com/wostzone/hub/auth/pkg/authorize"
	"github.com/wostzone/hub/lib/serve/pkg/certsetup"
)

// ACL filter function for authorization of thing access
// Todo: include groups the user is a member of to
type AclFilter struct {
	userID     string
	certOU     string // user OU when certificate authenticated
	authorizer authorize.VerifyAuthorization
}

// FilterThing returns true if user can read the Thing with ID thingID
// plugin certificates have full read access
func (aclFilter *AclFilter) FilterThing(thingID string) bool {
	if aclFilter.certOU == certsetup.OUPlugin {
		return true
	}
	if aclFilter.userID == "" || thingID == "" {
		return false
	}
	// authorize read access
	hasAccess := aclFilter.authorizer(aclFilter.userID, aclFilter.certOU, thingID, false, "")
	return hasAccess
}

// NewAclFilter. Provide authorization context needed to authorize requests
// userID to filter on. An empty userID always fails.
// authorizer is the function that performs the actual authorization
func NewAclFilter(userID string, certOU string, authorizer authorize.VerifyAuthorization) AclFilter {
	return AclFilter{
		authorizer: authorizer,
		userID:     userID,
		certOU:     certOU,
	}
}
