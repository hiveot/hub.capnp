# Cap'n proto definition for the authentication service
@0xc2f3c14cadbaf856;

using Go = import "/go.capnp";
$Go.package("hubapi");
$Go.import("github.com/hiveot/hub/api/go/hubapi");

const authnServiceName :Text = "authn";
# ServiceName of the service used for logging and connecting

const defaultAccessTokenValiditySec :Int32 = 3600;
# DefaultAccessTokenValiditySec with access token validity in seconds

const defaultRefreshTokenValiditySec : Int32 = 1209600;
# DefaultRefreshTokenValiditySec with Refresh token validity before refresh

struct UserProfile {
# Container for profile information

	loginID @0 :Text;
	# The user's login ID, typically email

	name @1 :Text;
	# The user's display name for presentation
}

const capNameUserAuthn :Text = "capUserAuthn";
# The capability name for user authentication

const capNameManageAuthn :Text = "capManageAuthn";
# The capability name for managing authentication

interface CapAuthn {
# CapAuthn defines the interface for simple user management and authentication

	capUserAuthn @0 (clientID :Text) -> (cap :CapUserAuthn);
	# CapUserAuthn provides the authentication capabilities for unauthenticated users.
	#  ClientID is the ID of the user of the capability

	capManageAuthn @1 (clientID :Text) -> (cap :CapManageAuthn);
	# CapManageAuthn provides the capability manage users for use by administrators.
	#  ClientID is the ID of the user of the capability
}


interface CapManageAuthn {
# CapManageAuthn defines the interface for managing the authentication service
# Intended for administrators only.

	addUser @0 (loginID :Text, password :Text) -> (password :Text);
	# AddUser adds a new user and generates a temporary password.
	# If the user already exists an error is returned.
	# Users can set their own name with CapUserAuthn.UpdateProfile.
	# Optionally provide a password or use "" to auto generate one

	listUsers @1() -> (profiles :List(UserProfile));
	# ListUsers provide a list of users and their info

	removeUser @2 (loginID :Text) -> ();
	# RemoveUser removes a user and disables login
	# Existing tokens are immediately expired (tbd)

	resetPassword @3 (loginID :Text, newPassword :Text) -> (password :Text);
	# ResetPassword reset the user's password and returns a new password
	# Optionally provide a new password or use "" to auto generate one
}

interface CapUserAuthn {
# CapAuthentication defines the capabilities to handle authentication of a client
# Intended for end-users to login, logout, or obtain their profile

	getProfile @0 () -> (profile :UserProfile);
	# GetProfile returns the user's profile after successful authentication

	login @1 (password :Text) -> (authToken :Text, refreshToken :Text);
	# Login to authenticate a user
	# This returns a short lived auth token for use with the HTTP api,
	# and a medium lived refresh token used to obtain a new auth token.

	logout @2 (refreshToken :Text) -> ();
	# Logout invalidates the refresh token for the user

	refresh @3 (refreshToken :Text) -> (newAuthToken :Text, newRefreshToken :Text);
	# Refresh an authentication token
	# refreshToken must be a valid refresh token obtained at login
	# This returns a short lived auth token and medium lived refresh token

	setPassword @4 (newPassword :Text) -> ();
	# SetPassword changes the client password after successful authentication

	setProfile @5 (newProfile :UserProfile) -> ();
	# SetProfile replaces the user profile after successful authentication
}
