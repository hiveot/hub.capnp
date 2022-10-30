package capserializer

import (
	"capnproto.org/go/capnp/v3"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub/pkg/authn"
)

func MarshalUserProfile(profile authn.UserProfile) (profileCapnp hubapi.UserProfile) {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	profileCapnp, _ = hubapi.NewUserProfile(seg)
	_ = profileCapnp.SetName(profile.Name)
	_ = profileCapnp.SetLoginID(profile.LoginID)
	return profileCapnp
}

func MarshalUserProfileList(profiles []authn.UserProfile) (
	profileListCapnp hubapi.UserProfile_List) {

	_, seg, _ := capnp.NewMessage(capnp.SingleSegment(nil))
	profileListCapnp, _ = hubapi.NewUserProfile_List(seg, int32(len(profiles)))
	for i, profile := range profiles {
		profileCapnp := MarshalUserProfile(profile)
		_ = profileListCapnp.Set(i, profileCapnp)
	}
	return profileListCapnp
}

func UnmarshalUserProfile(profileCapnp hubapi.UserProfile) (profile authn.UserProfile) {
	userName, _ := profileCapnp.Name()
	loginID, _ := profileCapnp.LoginID()
	profile.Name = userName
	profile.LoginID = loginID
	return profile
}

func UnmarshalUserProfileList(profileListCapnp hubapi.UserProfile_List) (profileList []authn.UserProfile) {
	count := profileListCapnp.Len()
	for i := 0; i < count; i++ {
		profileCapnp := profileListCapnp.At(i)
		profile := UnmarshalUserProfile(profileCapnp)
		profileList = append(profileList, profile)
	}
	return profileList
}
