package launcher_test

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
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/internal/svcconfig"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capnpclient"
	"github.com/hiveot/hub/pkg/launcher/capnpserver"
	"github.com/hiveot/hub/pkg/launcher/config"
	"github.com/hiveot/hub/pkg/launcher/service"
)

// when testing using the capnp RPC
const testAddress = "/tmp/launcher_test.socket"
const useTestCapnp = true

var homeFolder string = "/tmp"
var logFolder string = "/tmp"

func newServer(useCapnp bool) (launcher.ILauncher, func()) {
	//cwd, _ := os.Getwd()
	//homeFolder = path.Join(cwd, "../../dist")
	var launcherConfig = config.NewLauncherConfig()
	launcherConfig.AttachStderr = true
	launcherConfig.AttachStdout = false
	launcherConfig.LogServices = true
	var f = svcconfig.GetFolders(homeFolder, false)
	f.Services = "/bin" // for /bin/yes
	f.Logs = logFolder

	ctx, cancelFunc := context.WithCancel(context.Background())
	svc, err := service.NewLauncherService(ctx, f, launcherConfig)
	if err != nil {
		logrus.Fatal(err)
	}

	// optionally test with capnp RPC
	if useCapnp {
		_ = syscall.Unlink(testAddress)
		srvListener, _ := net.Listen("unix", testAddress)
		go capnpserver.StartLauncherCapnpServer(ctx, srvListener, svc)
		// connect the client to the server above
		clConn, _ := net.Dial("unix", testAddress)
		cl, err := capnpclient.NewLauncherCapnpClient(ctx, clConn)
		if err != nil {
			logrus.Fatalf("Failed starting capnp client: %s", err)
		}
		return cl, cancelFunc
	}
	return svc, cancelFunc
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	svc, cancelFunc := newServer(useTestCapnp)
	assert.NotNil(t, svc)

	cancelFunc()
}

func TestList(t *testing.T) {
	ctx := context.Background()
	svc, cancelFunc := newServer(useTestCapnp)
	require.NotNil(t, svc)
	info, err := svc.List(ctx, false)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	cancelFunc()
}

func TestStartYes(t *testing.T) {
	// remove logfile from previous run
	logFile := path.Join(logFolder, "yes.log")
	_ = os.Remove(logFile)

	//
	ctx := context.Background()
	svc, cancelFunc := newServer(useTestCapnp)
	assert.NotNil(t, svc)
	info, err := svc.Start(ctx, "yes")
	require.NoError(t, err)
	assert.True(t, info.Running)
	assert.True(t, info.PID > 0)
	assert.True(t, info.StartTime != "")
	assert.FileExists(t, logFile)

	info2, err := svc.Stop(ctx, "yes")
	time.Sleep(time.Millisecond * 10)
	assert.NoError(t, err)
	assert.False(t, info2.Running)
	assert.True(t, info2.StopTime != "")

	cancelFunc()
}

func TestStartBadName(t *testing.T) {
	ctx := context.Background()
	svc, cancelFunc := newServer(useTestCapnp)
	assert.NotNil(t, svc)

	_, err := svc.Start(ctx, "notaservicename")
	require.Error(t, err)
	//
	_, err = svc.Stop(ctx, "notaservicename")
	require.Error(t, err)
	//
	cancelFunc()
}

func TestStartStopTwice(t *testing.T) {
	ctx := context.Background()
	svc, cancelFunc := newServer(useTestCapnp)
	assert.NotNil(t, svc)

	info, err := svc.Start(ctx, "yes")
	assert.NoError(t, err)
	// again
	info2, err := svc.Start(ctx, "yes")
	assert.Error(t, err)
	_ = info2
	//assert.Equal(t, info.PID, info2.PID)

	// stop twice
	info3, err := svc.Stop(ctx, "yes")
	assert.NoError(t, err)
	assert.False(t, info3.Running)
	assert.Equal(t, info.PID, info3.PID)
	// stopping is idempotent
	info4, err := svc.Stop(ctx, "yes")
	assert.NoError(t, err)
	assert.False(t, info3.Running)
	assert.Equal(t, info.PID, info4.PID)

	cancelFunc()
}

func TestStartStopAll(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	svc, cancelFunc := newServer(useTestCapnp)
	assert.NotNil(t, svc)

	_, err := svc.Start(ctx, "yes")
	assert.NoError(t, err)

	// result should be 1 service running
	info, err := svc.List(ctx, true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(info))

	// stopping
	err = svc.StopAll(ctx)
	assert.NoError(t, err)

	// result should be no service running
	info, err = svc.List(ctx, true)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(info))

	cancelFunc()
}
