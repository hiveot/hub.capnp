package lib_test

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/pkg/lib"
)

func TestWaitForSignal(t *testing.T) {
	var waitCompleted = false
	go func() {
		lib.WaitForSignal()
		waitCompleted = true
	}()
	pid := os.Getpid()
	time.Sleep(time.Second)

	// signal.Notify()
	syscall.Kill(pid, syscall.SIGINT)
	time.Sleep(time.Second)

	assert.True(t, waitCompleted)
}
