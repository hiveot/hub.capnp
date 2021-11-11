package idprovclient_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/idprovclient"
)

const serviceID = "testID"
const port = 1234
const ipString = "127.0.0.1"

const serviceName = "test-idprov"

// ServeDiscovery serves a discovery record on localhost for testing
// map for parameters to include in the TXT record, or nil to exclude the TXT record
func ServeDiscovery(params map[string]string) (*zeroconf.Server, error) {
	serviceType := fmt.Sprintf("_%s._tcp", serviceName)
	domain := "local."
	hostname := "localhost"
	ips := []string{ipString}
	ifaces := []net.Interface{}
	textRecord := []string{}
	for k, v := range params {
		textRecord = append(textRecord, fmt.Sprintf("%s=%s", k, v))
	}

	server, err := zeroconf.RegisterProxy(
		serviceID, serviceType, domain, int(port), hostname, ips, textRecord, ifaces)
	return server, err
}

// Test discovery of a idprov server by publishing a discovery record
func TestDiscover(t *testing.T) {
	const path = "/path/to/service"

	// Serve a discovery record
	params := map[string]string{"path": path}
	server, err := ServeDiscovery(params)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Test if it is discovered
	url, err := idprovclient.DiscoverProvisioningServer(serviceName, 1)
	// DNS-SD adds a '.' after the domain. Should we use IP instead?
	expectedURL := fmt.Sprintf("https://%s:%d%s", ipString, port, path)
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)

	time.Sleep(time.Second * 3)
	server.Shutdown()
}

func TestDiscoverMissingPath(t *testing.T) {
	// const path = "/path/to/service"
	server, err := ServeDiscovery(nil)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Test if it is discovered
	// This test will fail if another instance is already running
	url, err := idprovclient.DiscoverProvisioningServer(idprovclient.IdprovServiceName, 1)
	assert.Error(t, err)
	assert.Empty(t, url)

	server.Shutdown()
}

func TestDiscoverNoServer(t *testing.T) {
	// Test if it is discovered
	_, err := idprovclient.DiscoverProvisioningServer(idprovclient.IdprovServiceName, 1)
	assert.Error(t, err)
}
