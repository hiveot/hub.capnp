package hubnet

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	tokenString := "bearertokenstring"
	req, _ := http.NewRequest("GET", "someurl", http.NoBody)
	req.Header.Add("Authorization", "Bearer "+tokenString)

	token, err := GetBearerToken(req)
	assert.NoError(t, err)
	assert.Equal(t, tokenString, token)
}

func TestNoBearerToken(t *testing.T) {
	req, _ := http.NewRequest("GET", "someurl", http.NoBody)
	_, err := GetBearerToken(req)
	assert.Error(t, err)
}

func TestBadBearerToken(t *testing.T) {
	req, _ := http.NewRequest("GET", "someurl", http.NoBody)
	req.Header.Add("Authorization", "Bearer: bad token")
	_, err := GetBearerToken(req)
	assert.Error(t, err)

	req, _ = http.NewRequest("GET", "someurl", http.NoBody)
	req.Header.Add("Authorization", "NotBearer: token")
	_, err = GetBearerToken(req)
	assert.Error(t, err)
}

func TestGetOutboundInterface(t *testing.T) {
	name, mac, addr := GetOutboundInterface("")
	assert.NotEmpty(t, name)
	assert.NotEmpty(t, mac)
	assert.NotEmpty(t, addr)
	logrus.Infof("TestGetOutboundInterface: name=%s, mac=%s, addr=%s", name, mac, addr)

	name, _, _ = GetOutboundInterface("badaddress")
	assert.Empty(t, name)
}

func TestGetOutboundIP(t *testing.T) {
	addr := GetOutboundIP("")
	assert.NotEmpty(t, addr)
	logrus.Infof("TestGetOutboundIP: localhost= %s", addr)

	addr = GetOutboundIP("badaddress")
	assert.Empty(t, addr)
}
