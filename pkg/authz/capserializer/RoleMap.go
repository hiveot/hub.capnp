package capserializer

import (
	"capnproto.org/go/capnp/v3"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authz"
)

// MarshalRoleMap convert role map from POGS to capnp format
func MarshalRoleMap(roles authz.RoleMap) hubapi.RoleMap {
	// rolemap is a list of K:V entries
	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	roleMapCapnp, _ := hubapi.NewRoleMap(seg) //, int32(len(roles)))

	_, seg2, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	roleMapEntriesCapnp, _ := hubapi.NewRoleMap_Entry_List(seg2, int32(len(roles)))

	i := 0
	for key, role := range roles {
		_, seg3, _ := capnp.NewMessage(capnp.SingleSegment(nil))
		roleEntryCapnp, _ := hubapi.NewRoleMap_Entry(seg3)
		//roleEntryCapnp := roleMapEntriesCapnp.At(i)
		roleEntryCapnp.SetKey(key)
		roleEntryCapnp.SetRole(role)
		roleMapEntriesCapnp.Set(i, roleEntryCapnp)
		i++
	}
	// WARNING: SetEntries must take place after roleMapEntriesCapnp is filled in.
	// if set after allocation but before filling in entries, it will be empty.
	roleMapCapnp.SetEntries(roleMapEntriesCapnp)
	return roleMapCapnp
}

func UnmarshalRoleMap(roleMapCapnp hubapi.RoleMap) authz.RoleMap {
	var roleKey string
	var roleID string
	var err error

	roleMapPOGS := make(authz.RoleMap)
	entriesCapnp, _ := roleMapCapnp.Entries()
	for i := 0; i < entriesCapnp.Len(); i++ {
		roleEntryCapnp := entriesCapnp.At(i)
		roleKey, err = roleEntryCapnp.Key()
		if err == nil {
			roleID, err = roleEntryCapnp.Role()
		}
		if err != nil {
			logrus.Errorf("conversion read entry failure: %s", err)
		}
		// clientRolePOGS := ClientRoleCapnp2POGS(clientRoleCapnp)
		roleMapPOGS[roleKey] = roleID
	}
	return roleMapPOGS
}
