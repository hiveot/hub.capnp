package hub_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/pkg/hub"
)

func TestStartHub(t *testing.T) {
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	err := hub.StartHub(homeFolder, false)
	assert.NoError(t, err)

	time.Sleep(3 * time.Second)
	hub.StopHub()
}
