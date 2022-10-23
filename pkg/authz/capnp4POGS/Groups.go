package capnp4POGS

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authz"
)

// GroupListPOGS2Capnp convert from POGS to capnp format
func GroupListPOGS2Capnp(groups []authz.Group) (groupListCapnp hubapi.Group_List) {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	groupListCapnp, _ = hubapi.NewGroup_List(seg, int32(len(groups)))
	for i, group := range groups {
		groupCapnp := GroupPOGS2Capnp(group)
		_ = groupListCapnp.Set(i, groupCapnp)
	}
	return groupListCapnp
}

// GroupListCapnp2POGS convert from capnp 2 POGS format
func GroupListCapnp2POGS(groupListCapnp hubapi.Group_List) (groups []authz.Group) {
	groups = make([]authz.Group, groupListCapnp.Len())
	for i := 0; i < groupListCapnp.Len(); i++ {
		groupCapnp := groupListCapnp.At(i)
		group := GroupCapnp2POGS(groupCapnp)
		groups[i] = group
	}
	return groups
}

func GroupPOGS2Capnp(group authz.Group) (groupCapnp hubapi.Group) {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	groupCapnp, _ = hubapi.NewGroup(seg)
	rolesCapnp := RoleMapPOGS2Capnp(group.MemberRoles)
	_ = groupCapnp.SetName(group.Name)
	_ = groupCapnp.SetMemberRoles(rolesCapnp)
	return groupCapnp
}

func GroupCapnp2POGS(groupCapnp hubapi.Group) (group authz.Group) {
	rolesCapnp, _ := groupCapnp.MemberRoles()
	name, _ := groupCapnp.Name()
	group.MemberRoles = RoleMapCapnp2POGS(rolesCapnp)
	group.Name = name
	return group
}
