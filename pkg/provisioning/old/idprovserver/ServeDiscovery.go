package idprovserver

import (
	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/wost-go/pkg/discovery"
)

// ServeIdProvDiscovery publishes a discovery record of the IDProv server
// Returns the discovery service instance. Use Shutdown() when done.
func (srv *IDProvServer) ServeIdProvDiscovery(serviceName string) (*zeroconf.Server, error) {
	params := map[string]string{"path": srv.directory.Endpoints.GetDirectory}

	directoryPath := srv.directory.Endpoints.GetDirectory

	logrus.Infof("ServeIdProvDiscovery serviceName=%s, service: %s:%d%s",
		serviceName, srv.config.IdpAddress, srv.config.IdpPort, directoryPath)

	return discovery.DiscoServe(srv.config.ClientID, serviceName, srv.config.IdpAddress, srv.config.IdpPort, params)

}
