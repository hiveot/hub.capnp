package capserializer

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub/api/go/hubapi"
	"github.com/hiveot/hub/pkg/authz"
)

// MarshalGroupList serializes a group list to a capnp message
func MarshalGroupList(groups []authz.Group) (groupListCapnp hubapi.Group_List) {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	groupListCapnp, _ = hubapi.NewGroup_List(seg, int32(len(groups)))
	for i, group := range groups {
		groupCapnp := MarshalGroup(group)
		_ = groupListCapnp.Set(i, groupCapnp)
	}
	return groupListCapnp
}

// UnmarshalGroupList deserializes a group list from a capnp message
func UnmarshalGroupList(groupListCapnp hubapi.Group_List) (groups []authz.Group) {
	groups = make([]authz.Group, groupListCapnp.Len())
	for i := 0; i < groupListCapnp.Len(); i++ {
		groupCapnp := groupListCapnp.At(i)
		group := UnmarshalGroup(groupCapnp)
		groups[i] = group
	}
	return groups
}

func MarshalGroup(group authz.Group) (groupCapnp hubapi.Group) {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	groupCapnp, _ = hubapi.NewGroup(seg)
	rolesCapnp := MarshalRoleMap(group.MemberRoles)
	_ = groupCapnp.SetName(group.Name)
	_ = groupCapnp.SetMemberRoles(rolesCapnp)
	return groupCapnp
}

func UnmarshalGroup(groupCapnp hubapi.Group) (group authz.Group) {
	rolesCapnp, _ := groupCapnp.MemberRoles()
	name, _ := groupCapnp.Name()
	group.MemberRoles = UnmarshalRoleMap(rolesCapnp)
	group.Name = name
	return group
}
