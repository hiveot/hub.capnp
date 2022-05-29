package dirserver

import (
	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/wost-go/pkg/discovery"
)

const ThingDirServiceDiscoveryType = "thingdir"

// ServeDirDiscovery publishes a discovery record of the directory server
// Returns the discovery service instance. Use Shutdown() when done.
func ServeDirDiscovery(instanceID string, serviceName string, address string, port uint) (*zeroconf.Server, error) {

	logrus.Infof("ServeDirDiscovery serviceID='%s;, address='%s:%d'", serviceName, address, port)

	return discovery.DiscoServe(instanceID, serviceName, address, port, nil)

}
