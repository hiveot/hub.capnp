package authz_test

import (
	"context"
	"net"
	"os"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/pkg/authz"
	"github.com/hiveot/hub/pkg/authz/capnp4POGS"
	"github.com/hiveot/hub/pkg/authz/capnpclient"
	"github.com/hiveot/hub/pkg/authz/capnpserver"
	"github.com/hiveot/hub/pkg/authz/service"
)

const aclFileName = "test-authz.acl" // auth_opt_aclFile
const testUseCapnp = true
const testAddress = "/tmp/authz_test.socket"

var aclFilePath string

var tempFolder string

// Create a new auth service with empty acl list
func startTestAuthzService(useCapnp bool) (svc authz.IAuthz, release func()) {

	ctx, cancelFunc := context.WithCancel(context.Background())
	_ = cancelFunc
	_ = os.Remove(aclFilePath)
	authSvc := service.NewAuthzService(ctx, aclFilePath)
	err := authSvc.Start(ctx)
	if err != nil {
		return nil, nil
	}
	if useCapnp {
		// start the capnp server
		_ = syscall.Unlink(testAddress)
		srvListener, err := net.Listen("unix", testAddress)
		if err != nil {
			logrus.Panic("Unable to create a listener, can't run test")
		}
		go capnpserver.StartAuthzCapnpServer(ctx, srvListener, authSvc)

		// connect the client to the server above
		clConn, _ := net.Dial("unix", testAddress)
		capClient, _ := capnpclient.NewAuthzCapnpClient(ctx, clConn)
		return capClient, cancelFunc
	}
	return authSvc, authSvc.Stop
}

// TestMain for all authn tests, setup of default folders and filenames
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	tempFolder = path.Join(os.TempDir(), "hiveot-authz-test")
	_ = os.MkdirAll(tempFolder, 0700)

	aclFilePath = path.Join(tempFolder, aclFileName)

	// creating these files takes a bit of time,
	time.Sleep(time.Second)

	res := m.Run()
	if res == 0 {
		_ = os.RemoveAll(tempFolder)
	}
	os.Exit(res)
}

func TestAuthzServiceStartStop(t *testing.T) {
	logrus.Infof("---TestAuthzServiceStartStop---")
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()

	time.Sleep(time.Second * 1)
	assert.NotNil(t, svc)
}

func TestAuthzServiceBadStart(t *testing.T) {
	logrus.Infof("---TestAuthzServiceBadStart---")
	ctx := context.Background()
	badAclFilePath := "/bad/aclstore/path"
	svc := service.NewAuthzService(ctx, badAclFilePath)

	// opening the acl store should fail
	err := svc.Start(ctx)
	assert.Error(t, err)
	svc.Stop()

	// missing store should not panic
	svc = service.NewAuthzService(ctx, "")
	err = svc.Start(ctx)
	assert.Error(t, err)
}

func TestIsPublisher(t *testing.T) {
	const testDevice1 = "device1ID"
	logrus.Infof("---TestIsPublisher---")
	thingID1 := "urn:zone:" + testDevice1 + ":sensor1:temperature"
	thingID2 := "urn:zone:" + testDevice1 + ":sensor1"
	thingID3 := "urn:zone:" + testDevice1 + ""

	// setup
	ctx := context.Background()
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()
	verifyAuthz := svc.CapVerifyAuthz(ctx)

	isPublisher, _ := verifyAuthz.IsPublisher(ctx, testDevice1, thingID1)
	assert.True(t, isPublisher)
	isPublisher, _ = verifyAuthz.IsPublisher(ctx, testDevice1, thingID2)
	assert.False(t, isPublisher)
	isPublisher, _ = verifyAuthz.IsPublisher(ctx, testDevice1, thingID3)
	assert.False(t, isPublisher)
}

// Test that devices have authorization to publish TDs and events
func TestDeviceAuthorization(t *testing.T) {
	logrus.Infof("---TestDeviceAuthorization---")
	const group1ID = "group1"
	const device1ID = "pub1"
	const thingID1 = "urn:zone1:pub1:device1:sensor1"
	const thingID2 = "urn:zone1:pub2:device2:sensor1"

	// setup
	ctx := context.Background()
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()
	verifyAuthz := svc.CapVerifyAuthz(ctx)
	manageAuthz := svc.CapManageAuthz(ctx)

	// FIXME: the device ID is normally not a member of the group
	err := manageAuthz.SetClientRole(ctx, device1ID, group1ID, authz.ClientRoleIotDevice)
	assert.NoError(t, err)
	err = manageAuthz.SetClientRole(ctx, thingID1, group1ID, authz.ClientRoleIotDevice)
	assert.NoError(t, err)
	err = manageAuthz.SetClientRole(ctx, thingID2, group1ID, authz.ClientRoleIotDevice)
	assert.NoError(t, err)

	// this test makes no sense as devices have authz but are not in ACLs
	perms, _ := verifyAuthz.GetPermissions(ctx, device1ID, thingID1)
	assert.Contains(t, perms, authz.PermPubTD)
	assert.Contains(t, perms, authz.PermPubEvent)
	assert.Contains(t, perms, authz.PermReadAction)
	assert.NotContains(t, perms, authz.PermWriteProperty)
	assert.NotContains(t, perms, authz.PermEmitAction)
}

func TestManagerAuthorization(t *testing.T) {
	logrus.Infof("---TestManagerAuthorization---")
	const client1ID = "manager1"
	const group1ID = "group1"
	const thingID1 = "urn:zone1:pub1:device1:sensor1"
	const thingID2 = "urn:zone1:pub2:device1:sensor1"

	// setup
	ctx := context.Background()
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()

	verifyAuthz := svc.CapVerifyAuthz(ctx)
	manageAuthz := svc.CapManageAuthz(ctx)
	_ = manageAuthz.SetClientRole(ctx, thingID1, group1ID, authz.ClientRoleIotDevice)
	_ = manageAuthz.SetClientRole(ctx, thingID2, group1ID, authz.ClientRoleIotDevice)

	// services can do whatever as a manager in the all group
	// the manager in the allgroup takes precedence over the operator role in group1
	_ = manageAuthz.SetClientRole(ctx, client1ID, group1ID, authz.ClientRoleOperator)
	_ = manageAuthz.SetClientRole(ctx, client1ID, authz.AllGroupName, authz.ClientRoleManager)
	perms, _ := verifyAuthz.GetPermissions(ctx, client1ID, thingID1)

	assert.Contains(t, perms, authz.PermReadTD)
	assert.Contains(t, perms, authz.PermReadEvent)
	assert.Contains(t, perms, authz.PermEmitAction)
	assert.Contains(t, perms, authz.PermWriteProperty)
	assert.NotContains(t, perms, authz.PermPubTD)
	assert.NotContains(t, perms, authz.PermPubEvent)

}

func TestOperatorAuthorization(t *testing.T) {
	logrus.Infof("---TestOperatorAuthorization---")
	const client1ID = "operator1"
	const deviceID = "device1"
	const group1ID = "group1"
	const thingID1 = "urn:zone1:pub1:device1:sensor1"
	const thingID2 = "urn:zone1:pub2:device1:sensor1"
	ctx := context.Background()

	// setup
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()

	manageAuthz := svc.CapManageAuthz(ctx)
	verifyAuthz := svc.CapVerifyAuthz(ctx)
	err := manageAuthz.SetClientRole(ctx, thingID1, group1ID, authz.ClientRoleIotDevice)
	assert.NoError(t, err)
	_ = manageAuthz.SetClientRole(ctx, thingID2, group1ID, authz.ClientRoleIotDevice)

	_ = manageAuthz.SetClientRole(ctx, deviceID, group1ID, authz.ClientRoleIotDevice)
	_ = manageAuthz.SetClientRole(ctx, client1ID, group1ID, authz.ClientRoleOperator)

	// operators can readTD, readEvent, emitAction
	_ = manageAuthz.SetClientRole(ctx, client1ID, group1ID, authz.ClientRoleOperator)
	perms, _ := verifyAuthz.GetPermissions(ctx, client1ID, thingID1)

	assert.Contains(t, perms, authz.PermReadTD)
	assert.Contains(t, perms, authz.PermReadEvent)
	assert.Contains(t, perms, authz.PermEmitAction)
	assert.NotContains(t, perms, authz.PermPubEvent)
	assert.NotContains(t, perms, authz.PermPubTD)
	assert.NotContains(t, perms, authz.PermWriteProperty)

}

func TestViewerAuthorization(t *testing.T) {
	logrus.Infof("---TestViewerAuthorization---")
	const user1ID = "viewer1"
	const group1ID = "group1"
	const thingID1 = "urn:zone1:pub1:device1:sensor1"
	const thingID2 = "urn:zone1:pub2:device1:sensor1"

	// setup
	ctx := context.Background()
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()

	verifyAuthz := svc.CapVerifyAuthz(ctx)
	manageAuthz := svc.CapManageAuthz(ctx)
	err := manageAuthz.SetClientRole(ctx, thingID1, group1ID, authz.ClientRoleIotDevice)
	assert.NoError(t, err)
	_ = manageAuthz.SetClientRole(ctx, thingID2, group1ID, authz.ClientRoleIotDevice)

	// viewers role can read TD
	_ = manageAuthz.SetClientRole(ctx, user1ID, group1ID, authz.ClientRoleViewer)
	perms, _ := verifyAuthz.GetPermissions(ctx, user1ID, thingID1)

	assert.Contains(t, perms, authz.PermReadTD)
	assert.Contains(t, perms, authz.PermReadEvent)
	assert.NotContains(t, perms, authz.PermEmitAction)
	assert.NotContains(t, perms, authz.PermPubEvent)
	assert.NotContains(t, perms, authz.PermPubTD)
	assert.NotContains(t, perms, authz.PermWriteProperty)
}

func TestNoAuthorization(t *testing.T) {
	logrus.Infof("---TestNoAuthorization---")
	const user1ID = "viewer1"
	const group1ID = "group1"
	const thingID1 = "urn:zone1:pub1:device1:sensor1"
	const thingID2 = "urn:zone1:pub2:device1:sensor1"

	// setup
	ctx := context.Background()
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()

	verifyAuthz := svc.CapVerifyAuthz(ctx)
	manageAuthz := svc.CapManageAuthz(ctx)
	err := manageAuthz.AddThing(ctx, thingID1, group1ID)
	assert.NoError(t, err)
	_ = manageAuthz.AddThing(ctx, thingID2, group1ID)

	// viewers role can read TD
	_ = manageAuthz.SetClientRole(ctx, user1ID, group1ID, "badrole")
	perms, _ := verifyAuthz.GetPermissions(ctx, user1ID, thingID1)
	assert.Equal(t, 0, len(perms))
}

func TestListGroups(t *testing.T) {
	const user1ID = "viewer1"
	const group1ID = "group1"
	const group2ID = "group2"
	const group3ID = "group3"
	const thingID1 = "urn:pub1:device1:sensor1"
	const thingID2 = "urn:pub2:device2:sensor2"
	const thingID3 = "urn:pub2:device3:sensor1"

	// setup
	ctx := context.Background()
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()

	manageAuthz := svc.CapManageAuthz(ctx)
	err := manageAuthz.AddThing(ctx, thingID1, group1ID)
	assert.NoError(t, err)
	_ = manageAuthz.AddThing(ctx, thingID1, group2ID)
	_ = manageAuthz.AddThing(ctx, thingID2, group2ID)
	_ = manageAuthz.AddThing(ctx, thingID3, group3ID)
	_ = manageAuthz.SetClientRole(ctx, user1ID, group1ID, authz.ClientRoleViewer)
	_ = manageAuthz.SetClientRole(ctx, user1ID, group2ID, authz.ClientRoleViewer)

	// 3 groups must exist
	groups, err := manageAuthz.ListGroups(ctx, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(groups))

	// group 2 has 3 members, 2 things and 1 user
	group, err := manageAuthz.GetGroup(ctx, group2ID)
	assert.NoError(t, err)
	assert.Equal(t, group2ID, group.Name)
	assert.Equal(t, 3, len(group.MemberRoles))
	assert.Contains(t, group.MemberRoles, thingID1)
	assert.Contains(t, group.MemberRoles, thingID2)
	assert.Contains(t, group.MemberRoles, user1ID)

	// viewer1 is a member of 2 groups
	roles, err := manageAuthz.GetGroupRoles(ctx, thingID1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(roles))
	assert.Contains(t, roles, group1ID)
	assert.Contains(t, roles, group2ID)

	// a non existing group has no members
	group, err = manageAuthz.GetGroup(ctx, "notagroup")
	assert.Error(t, err)
}

func TestRoleMap4Capnp(t *testing.T) {
	const user1ID = "user1"
	const user2ID = "user2"
	const role1 = "role1"
	const role2 = "role2"

	roles := make(authz.RoleMap)
	roles[user1ID] = role1
	roles[user2ID] = role2
	roleMapCapnp := capnp4POGS.RoleMapPOGS2Capnp(roles)
	// and back
	roles2 := capnp4POGS.RoleMapCapnp2POGS(roleMapCapnp)
	assert.Len(t, roles2, len(roles))
	assert.Equal(t, role1, roles[user1ID])
	assert.Equal(t, role2, roles[user2ID])
}

func TestAddRemoveRoles(t *testing.T) {
	const user1ID = "viewer1"
	const group1ID = "group1"
	const group2ID = "group2"
	const group3ID = "group3"
	const thingID1 = "urn:pub1:device1:sensor1"
	const thingID2 = "urn:pub2:device2:sensor2"
	//const thingID3 = "urn:pub2:device3:sensor1"
	// setup
	ctx := context.Background()
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()
	manageAuthz := svc.CapManageAuthz(ctx)

	// user1 is a member of 3 groups
	err := manageAuthz.SetClientRole(ctx, user1ID, group1ID, authz.ClientRoleOperator)
	assert.NoError(t, err)
	_ = manageAuthz.SetClientRole(ctx, user1ID, group2ID, authz.ClientRoleOperator)
	_ = manageAuthz.SetClientRole(ctx, user1ID, group3ID, authz.ClientRoleOperator)

	// thing1 is a member of 3 groups
	// adding a thing twice should not fail
	err = manageAuthz.AddThing(ctx, thingID1, group1ID)
	assert.NoError(t, err)
	_ = manageAuthz.AddThing(ctx, thingID1, group1ID)
	_ = manageAuthz.AddThing(ctx, thingID1, group2ID)
	_ = manageAuthz.AddThing(ctx, thingID1, group3ID)
	roles, err := manageAuthz.GetGroupRoles(ctx, thingID1)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(roles))

	// verify remove thing1 from group 2
	err = manageAuthz.RemoveThing(ctx, thingID1, group2ID)
	assert.NoError(t, err)
	group2, err := manageAuthz.GetGroup(ctx, group2ID)
	assert.NoError(t, err)
	assert.NotContains(t, group2.MemberRoles, thingID1)

	roles, err = manageAuthz.GetGroupRoles(ctx, thingID1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(roles))
	assert.NotContains(t, roles, group2ID)

	// remove is idempotent.
	err = manageAuthz.RemoveThing(ctx, thingID1, group2ID)
	assert.NoError(t, err)
	// thingID2 is not a member
	err = manageAuthz.RemoveThing(ctx, thingID2, group2ID)
	assert.NoError(t, err)
	err = manageAuthz.RemoveClient(ctx, thingID2, group2ID)
	assert.NoError(t, err)
	err = manageAuthz.RemoveClient(ctx, thingID2, "notagroup")
	assert.Error(t, err)

	// removing all should remove user from all groups
	err = manageAuthz.RemoveAll(ctx, user1ID)
	roles, err = manageAuthz.GetGroupRoles(ctx, user1ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(roles))
}

func TestClientPermissions(t *testing.T) {
	const user1ID = "user1"
	const group1ID = "group1"
	const group2ID = "group2"
	const group3ID = "group3"
	const thing1ID = "urn:thing1"
	// setup
	ctx := context.Background()
	svc, stopFn := startTestAuthzService(testUseCapnp)
	defer stopFn()

	clientAuthz := svc.CapClientAuthz(ctx, user1ID)
	manageAuthz := svc.CapManageAuthz(ctx)
	_ = manageAuthz.AddThing(ctx, thing1ID, group1ID)
	_ = manageAuthz.AddThing(ctx, thing1ID, group2ID)
	_ = manageAuthz.AddThing(ctx, thing1ID, group3ID)
	_ = manageAuthz.SetClientRole(ctx, user1ID, group1ID, authz.ClientRoleViewer)
	_ = manageAuthz.SetClientRole(ctx, user1ID, group2ID, authz.ClientRoleManager)
	_ = manageAuthz.SetClientRole(ctx, user1ID, group3ID, authz.ClientRoleOperator)

	// as a manager, permissions to read and emit actions
	perms, err := clientAuthz.GetPermissions(ctx, thing1ID)
	assert.NoError(t, err)
	assert.Contains(t, perms, authz.PermEmitAction)
	assert.Contains(t, perms, authz.PermWriteProperty)

	// after removing the manager role write property permissions no longer apply
	manageAuthz.RemoveThing(ctx, user1ID, group2ID)
	perms, err = clientAuthz.GetPermissions(ctx, thing1ID)
	assert.NoError(t, err)
	assert.Contains(t, perms, authz.PermEmitAction)
	assert.NotContains(t, perms, authz.PermWriteProperty)

}
