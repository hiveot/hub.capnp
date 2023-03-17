package listener_test

import (
	"github.com/hiveot/hub/lib/listener"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetOutboundInterface(t *testing.T) {
	name, mac, addr := listener.GetOutboundInterface("")
	assert.NotEmpty(t, name)
	assert.NotEmpty(t, mac)
	assert.NotEmpty(t, addr)
	logrus.Infof("TestGetOutboundInterface: name=%s, mac=%s, addr=%s", name, mac, addr)

	name, _, _ = listener.GetOutboundInterface("badaddress")
	assert.Empty(t, name)
}

func TestGetOutboundIP(t *testing.T) {
	addr := listener.GetOutboundIP("")
	assert.NotEmpty(t, addr)
	logrus.Infof("TestGetOutboundIP: localhost= %s", addr)

	addr = listener.GetOutboundIP("badaddress")
	assert.Empty(t, addr)
}
