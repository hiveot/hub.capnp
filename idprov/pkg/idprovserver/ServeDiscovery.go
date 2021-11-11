package idprovserver

import (
	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/serve/pkg/discovery"
)

const IdProvServiceName = "idprov"

// ServeIdProvDiscovery publishes a discovery record of the IDProv server
// Returns the discovery service instance. Use Shutdown() when done.
func (srv *IDProvServer) ServeIdProvDiscovery(serviceName string) (*zeroconf.Server, error) {
	params := map[string]string{"path": srv.directory.Endpoints.GetDirectory}

	directoryPath := srv.directory.Endpoints.GetDirectory

	logrus.Infof("ServeIdProvDiscovery serviceID=%s, service: %s:%d%s",
		srv.instanceID, srv.address, srv.port, directoryPath)

	return discovery.ServeDiscovery(srv.instanceID, serviceName, srv.address, srv.port, params)

}
