package authn_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/pkg/authn"
	"github.com/hiveot/hub/pkg/authn/capnpclient"
	"github.com/hiveot/hub/pkg/authn/capnpserver"
	"github.com/hiveot/hub/pkg/authn/config"
	"github.com/hiveot/hub/pkg/authn/service"
)

var passwordFile string // set in TestMain
const testUseCapnp = true

var tempFolder string
var testuser1 = "testuser1"
var testpass1 = "secret11" // set at start

// create a new authn service and set the password for testuser1
// containing a password for testuser1
func startTestAuthnService(useCapnp bool) (authSvc authn.IAuthnService, stopFn func(), err error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	_ = os.Remove(passwordFile)
	cfg := config.AuthnConfig{
		PasswordFile:            passwordFile,
		AccessTokenValiditySec:  10,
		RefreshTokenValiditySec: 120,
	}
	svc := service.NewAuthnService(cfg)
	err = svc.Start(ctx)
	if err == nil {
		mng, err2 := svc.CapManageAuthn(ctx, "test")
		err = err2
		defer mng.Release()
		testpass1, err = mng.AddUser(ctx, testuser1)
	}
	if err != nil {
		logrus.Panicf("cant start test authn service: %s", err)
	}
	if useCapnp {
		// start the server
		srvListener, err := net.Listen("tcp", ":0")
		if err != nil {
			logrus.Panic("Unable to create a listener, can't run test")
		}
		go capnpserver.StartAuthnCapnpServer(svc, srvListener)
		time.Sleep(time.Millisecond)

		// connect the client to the server above
		clConn, err := net.Dial("tcp", srvListener.Addr().String())
		capClient := capnpclient.NewAuthnCapnpClientConnection(ctx, clConn)
		return capClient, func() {
			capClient.Release()
			_ = clConn.Close()
			time.Sleep(time.Millisecond)
			cancelFunc()
			_ = svc.Stop()
		}, err
	}
	return svc, func() {
		cancelFunc()
		_ = svc.Stop()
	}, err
}

// TestMain creates a test environment
// Used for all test cases in this package
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	// a working folder for the data
	tempFolder = path.Join(os.TempDir(), "test-authn")
	_ = os.MkdirAll(tempFolder, 0700)

	// the password file to use
	passwordFile = path.Join(tempFolder, "test.passwd")

	res := m.Run()

	time.Sleep(time.Second)
	if res == 0 {
		_ = os.RemoveAll(tempFolder)
	}
	os.Exit(res)
}

// Create and verify a JWT token
func TestStartStop(t *testing.T) {
	srv, stopFn, err := startTestAuthnService(testUseCapnp)
	require.NoError(t, err)
	assert.NotNil(t, srv)
	time.Sleep(time.Millisecond * 10)
	stopFn()
}

// Create and verify a JWT token
func TestStartTwice(t *testing.T) {
	svc, stopFn, err := startTestAuthnService(testUseCapnp)
	require.NoError(t, err)
	require.NotNil(t, svc)

	stopFn()
}

// Create manage users
func TestManageUser(t *testing.T) {

	ctx := context.Background()
	svc, stopFn, err := startTestAuthnService(testUseCapnp)
	defer stopFn()
	require.NoError(t, err)
	mng, err := svc.CapManageAuthn(ctx, "test")
	assert.NoError(t, err)
	defer mng.Release()

	// expect the test user
	userList, err := mng.ListUsers(ctx)
	assert.NoError(t, err)
	require.Equal(t, 1, len(userList))

	profile1 := userList[0]
	assert.Equal(t, testuser1, profile1.LoginID)

	// remove user
	err = mng.RemoveUser(ctx, testuser1)
	assert.NoError(t, err)
	userList, err = mng.ListUsers(ctx)
	assert.NoError(t, err)
	require.Equal(t, 0, len(userList))

	// reset password adds the user again
	newpw, err := mng.ResetPassword(ctx, testuser1)
	assert.NoError(t, err)
	assert.NotEmpty(t, newpw)
	userList, err = mng.ListUsers(ctx)
	assert.NoError(t, err)
	require.Equal(t, 1, len(userList))

	// add existing user should fail
	_, err = mng.AddUser(ctx, testuser1)
	assert.Error(t, err)

}

func TestLoginRefreshLogout(t *testing.T) {
	var at1 string
	var rt1 string
	var at2 string
	var rt2 string
	count := 100
	ctx := context.Background()
	svc, stopFn, err := startTestAuthnService(testUseCapnp)
	defer stopFn()
	require.NoError(t, err)

	// login and get tokens
	clauth, err := svc.CapUserAuthn(ctx, testuser1)
	assert.NoError(t, err)
	defer clauth.Release()
	t1 := time.Now()
	for i := 0; i < count; i++ {
		at1, rt1, err = clauth.Login(ctx, testpass1)
	}
	d1 := time.Now().Sub(t1)
	assert.NoError(t, err)
	assert.NotEmpty(t, at1)
	assert.NotEmpty(t, rt1)

	// refresh token
	t2 := time.Now()
	for i := 0; i < count; i++ {
		at2, rt2, err = clauth.Refresh(ctx, rt1)
	}
	d2 := time.Now().Sub(t2)
	fmt.Printf("Time to login   %d times: %d msec\n", count, d1.Milliseconds())
	fmt.Printf("Time to refresh %d times: %d msec\n", count, d2.Milliseconds())
	assert.NoError(t, err)
	assert.NotEmpty(t, at2)
	assert.NotEmpty(t, rt2)

	// logout
	err = clauth.Logout(ctx, rt2)
	assert.NoError(t, err)
	// second logout should not give an error
	err = clauth.Logout(ctx, rt2)
	assert.NoError(t, err)
}

func TestLoginFail(t *testing.T) {
	ctx := context.Background()
	svc, stopFn, err := startTestAuthnService(testUseCapnp)
	defer stopFn()
	require.NoError(t, err)

	// login and get tokens
	clauth, err := svc.CapUserAuthn(ctx, testuser1)
	assert.NoError(t, err)
	defer clauth.Release()
	accessToken, refreshToken, err := clauth.Login(ctx, "badpass")
	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestProfile(t *testing.T) {
	ctx := context.Background()
	svc, stopFn, err := startTestAuthnService(testUseCapnp)
	defer stopFn()
	require.NoError(t, err)
	clauth, err := svc.CapUserAuthn(ctx, testuser1)
	assert.NoError(t, err)
	defer clauth.Release()

	// unauthenticated users cannot get their profile or set their password
	_, err = clauth.GetProfile(ctx)
	assert.Error(t, err)
	err = clauth.SetPassword(ctx, "passwordnotauthenticated")
	assert.Error(t, err)
	dummy := authn.UserProfile{}
	err = clauth.SetProfile(ctx, dummy)
	assert.Error(t, err)

	// after authentication get/set profile and get password should succeed
	at, rt, err := clauth.Login(ctx, testpass1)
	assert.NoError(t, err)
	assert.NotEmpty(t, at)
	assert.NotEmpty(t, rt)

	prof1, err := clauth.GetProfile(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testuser1, prof1.LoginID)

	prof1.Name = "new name"
	err = clauth.SetProfile(ctx, prof1)
	assert.NoError(t, err)
	err = clauth.SetPassword(ctx, "newpass")
	assert.NoError(t, err)

	// changing loginID in profile is not allowed
	prof1.LoginID = "new login"
	err = clauth.SetProfile(ctx, prof1)
	assert.Error(t, err)
}
