package listener_test

import (
	"fmt"
	"github.com/hiveot/hub/lib/hubclient"
	"github.com/hiveot/hub/lib/listener"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testServiceName = "test-service"
const testServiceType = "test-type"
const testServicePath = "/discovery/path"
const testServicePort = 9999

// Test the discovery client and server
func TestDiscover(t *testing.T) {
	testServiceAddress := listener.GetOutboundIP("").String()

	ServeDiscovery, err := listener.ServeDiscovery(
		testServiceName, testServiceType, testServiceAddress, testServicePort, 0, testServicePath)

	assert.NoError(t, err)
	assert.NotNil(t, ServeDiscovery)

	// Test if it is discovered
	address, port, discoParams, records, err := hubclient.DiscoverService(testServiceType, 1)
	require.NoError(t, err)
	rec0 := records[0]
	hostName, _ := os.Hostname()
	instanceID := testServiceName + "-" + hostName

	assert.Equal(t, instanceID, rec0.Instance)
	assert.Equal(t, testServiceAddress, address)
	assert.Equal(t, testServicePort, port)
	assert.Equal(t, testServicePath, discoParams["path"])

	time.Sleep(time.Millisecond) // prevent race error in discovery.server
	ServeDiscovery.Shutdown()
}

func TestDiscoverBadPort(t *testing.T) {
	badPort := 0
	address := listener.GetOutboundIP("").String()
	_, err := listener.ServeDiscovery(
		testServiceName, testServiceType, address, badPort, 0, "")

	assert.Error(t, err)
}

func TestNoServiceName(t *testing.T) {
	address := listener.GetOutboundIP("").String()

	_, err := listener.ServeDiscovery(
		"", testServiceType, address, testServicePort, 0, "")
	assert.Error(t, err) // missing instance name
}

func TestDiscoverNotFound(t *testing.T) {
	serviceName := "idprov-test"
	address := listener.GetOutboundIP("").String()

	ServeDiscoveryr, err := listener.ServeDiscovery(
		serviceName, testServiceType, address, testServicePort, 0, "")

	assert.NoError(t, err)

	// Test if it is discovered
	discoAddress, discoPort, _, records, err := hubclient.DiscoverService("wrongname", 1)
	_ = discoAddress
	_ = discoPort
	_ = records
	assert.Error(t, err)

	time.Sleep(time.Millisecond) // prevent race error in discovery.server
	ServeDiscoveryr.Shutdown()
	assert.Error(t, err)
}

func TestBadAddress(t *testing.T) {
	ServeDiscoveryr, err := listener.ServeDiscovery(
		testServiceName, testServiceType, "notanipaddress", testServicePort, 0, "")

	assert.Error(t, err)
	assert.Nil(t, ServeDiscoveryr)
}

func TestExternalAddress(t *testing.T) {
	ServeDiscoveryr, err := listener.ServeDiscovery(
		testServiceName, testServiceType, "1.2.3.4", testServicePort, 0, "")

	// expect a warning
	assert.NoError(t, err)
	time.Sleep(time.Millisecond) // prevent race error in discovery.server
	ServeDiscoveryr.Shutdown()
}

func TestDNSSDScan(t *testing.T) {

	records, err := hubclient.DnsSDScan("", 2)
	fmt.Printf("Found %d records in scan", len(records))

	assert.NoError(t, err)
	assert.Greater(t, len(records), 0, "No DNS records found")
}
