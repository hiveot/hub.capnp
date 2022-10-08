package launcher_test

import (
	"context"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub/pkg/launcher"
	"github.com/hiveot/hub/pkg/launcher/capnpclient"
	"github.com/hiveot/hub/pkg/launcher/capnpserver"
	"github.com/hiveot/hub/pkg/launcher/service"
)

// when testing using the capnp RPC
const testAddress = "/tmp/launcher_test.socket"
const useTestCapnp = true
const serviceFolder = "/bin"

func newServer(useCapnp bool) (launcher.ILauncher, func()) {
	svc := service.NewLauncherService(serviceFolder)
	ctx, cancelFunc := context.WithCancel(context.Background())
	_ = ctx

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
	assert.NotNil(t, svc)
	info, err := svc.List(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	cancelFunc()
}

func TestStartYes(t *testing.T) {
	ctx := context.Background()
	svc, cancelFunc := newServer(useTestCapnp)
	assert.NotNil(t, svc)
	info, err := svc.Start(ctx, "yes")
	require.NoError(t, err)
	assert.True(t, info.Running)
	assert.True(t, info.PID > 0)
	assert.True(t, info.StartTime != "")

	info2, err := svc.Stop(ctx, "yes")
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
	assert.Equal(t, info.PID, info2.PID)

	// stop twice
	info3, err := svc.Stop(ctx, "yes")
	assert.NoError(t, err)
	assert.Equal(t, info.PID, info3.PID)
	//
	info4, err := svc.Stop(ctx, "yes")
	assert.Error(t, err)
	assert.Equal(t, info.PID, info4.PID)

	cancelFunc()
}
