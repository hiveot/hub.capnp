package gateway_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wostzone/gateway/pkg/gateway"
)

func TestStartGateway(t *testing.T) {
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	err := gateway.StartGateway(homeFolder, false)
	assert.NoError(t, err)

	time.Sleep(3 * time.Second)
	gateway.StopGateway()
}
