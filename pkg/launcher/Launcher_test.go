package launcher_test

import (
	"context"
	"github.com/hiveot/hub/lib/hubclient"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub/lib/svcconfig"

	"github.com/hiveot/hub/lib/logging"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capnpclient"
	"github.com/hiveot/hub/pkg/launcher/capnpserver"
	"github.com/hiveot/hub/pkg/launcher/config"
	"github.com/hiveot/hub/pkg/launcher/service"
)

// when testing using the capnp RPC
const testUseCapnp = true

var homeFolder = "/tmp"
var logFolder = "/tmp"

func newServer(useCapnp bool) (l launcher.ILauncher, stopFn func()) {
	var launcherConfig = config.NewLauncherConfig()
	launcherConfig.AttachStderr = true
	launcherConfig.AttachStdout = false
	launcherConfig.LogServices = true
	var f = svcconfig.GetFolders(homeFolder, false)
	f.Services = "/bin" // for /bin/yes
	f.Bindings = ""
	f.Logs = logFolder

	ctx, cancelFunc := context.WithCancel(context.Background())
	svc := service.NewLauncherService(f, launcherConfig)
	err := svc.Start(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	// optionally test with capnp RPC
	if useCapnp {
		srvListener, _ := net.Listen("tcp", ":0")
		go capnpserver.StartLauncherCapnpServer(srvListener, svc)

		// connect the client to the server above
		capClient, _ := hubclient.ConnectWithCapnpTCP(srvListener.Addr().String(), nil, nil)
		cl, err := capnpclient.NewLauncherCapnpClient(capClient)
		if err != nil {
			logrus.Fatalf("Failed starting capnp client: %s", err)
		}
		return cl, func() {
			cl.Release()
			_ = srvListener.Close()
			cancelFunc()
			_ = svc.StopAll(ctx)
		}
	}
	return svc, func() {
		cancelFunc()
		_ = svc.StopAll(ctx)
	}
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	svc, cancelFunc := newServer(testUseCapnp)
	defer cancelFunc()
	assert.NotNil(t, svc)
}

func TestList(t *testing.T) {
	ctx := context.Background()
	svc, cancelFunc := newServer(testUseCapnp)
	defer cancelFunc()
	require.NotNil(t, svc)
	info, err := svc.List(ctx, false)
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestStartYes(t *testing.T) {
	// remove logfile from previous run
	logFile := path.Join(logFolder, "yes.log")
	_ = os.Remove(logFile)

	//
	ctx := context.Background()
	svc, cancelFunc := newServer(testUseCapnp)
	defer cancelFunc()

	assert.NotNil(t, svc)
	info, err := svc.StartService(ctx, "yes")
	require.NoError(t, err)
	assert.True(t, info.Running)
	assert.True(t, info.PID > 0)
	assert.True(t, info.StartTime != "")
	assert.FileExists(t, logFile)

	info2, err := svc.StopService(ctx, "yes")
	time.Sleep(time.Millisecond * 10)
	assert.NoError(t, err)
	assert.False(t, info2.Running)
	assert.True(t, info2.StopTime != "")
}

func TestStartBadName(t *testing.T) {
	ctx := context.Background()
	svc, cancelFunc := newServer(testUseCapnp)
	defer cancelFunc()
	assert.NotNil(t, svc)

	_, err := svc.StartService(ctx, "notaservicename")
	require.Error(t, err)
	//
	_, err = svc.StopService(ctx, "notaservicename")
	require.Error(t, err)
}

func TestStartStopTwice(t *testing.T) {
	ctx := context.Background()
	svc, cancelFunc := newServer(testUseCapnp)
	defer cancelFunc()
	assert.NotNil(t, svc)

	info, err := svc.StartService(ctx, "yes")
	assert.NoError(t, err)
	// again
	info2, err := svc.StartService(ctx, "yes")
	assert.Error(t, err)
	_ = info2
	//assert.Equal(t, info.PID, info2.PID)

	// stop twice
	info3, err := svc.StopService(ctx, "yes")
	assert.NoError(t, err)
	assert.False(t, info3.Running)
	assert.Equal(t, info.PID, info3.PID)
	// stopping is idempotent
	info4, err := svc.StopService(ctx, "yes")
	assert.NoError(t, err)
	assert.False(t, info3.Running)
	assert.Equal(t, info.PID, info4.PID)
}

func TestStartStopAll(t *testing.T) {
	ctx := context.Background()
	svc, cancelFunc := newServer(testUseCapnp)
	defer cancelFunc()
	assert.NotNil(t, svc)

	_, err := svc.StartService(ctx, "yes")
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

}
