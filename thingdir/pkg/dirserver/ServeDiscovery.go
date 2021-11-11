package dirserver

import (
	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/serve/pkg/discovery"
)

const ThingDirServiceDiscoveryType = "thingdir"

// ServeDirDiscovery publishes a discovery record of the IDProv server
// Returns the discovery service instance. Use Shutdown() when done.
func ServeDirDiscovery(instanceID string, serviceName string, address string, port uint) (*zeroconf.Server, error) {

	logrus.Infof("ServeDirDiscovery serviceID='%s;, address='%s:%d'", serviceName, address, port)

	return discovery.ServeDiscovery(instanceID, serviceName, address, port, nil)

}
