package gateway_test

import (
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/pkg/gateway"
)

func TestWaitForSignal(t *testing.T) {
	m := sync.Mutex{}
	var waitCompleted = false
	go func() {
		gateway.WaitForSignal()
		m.Lock()
		waitCompleted = true
		m.Unlock()
	}()
	pid := os.Getpid()
	time.Sleep(time.Second)

	// signal.Notify()
	syscall.Kill(pid, syscall.SIGINT)
	time.Sleep(time.Second)
	m.Lock()
	defer m.Unlock()
	assert.True(t, waitCompleted)
}
