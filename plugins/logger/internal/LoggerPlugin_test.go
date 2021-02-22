package internal_test

import (
	"testing"

	"github.com/wostzone/gateway/plugins/logger/internal"
)

func TestStartLoggerPlugin(t *testing.T) {
	config := &internal.Config{}
	lp := internal.NewLoggerPlugin(config)
	lp.Start()
}
