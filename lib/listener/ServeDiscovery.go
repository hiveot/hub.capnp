// Package service to publish Hub gateway for discovery
package listener

import (
	"fmt"
	"github.com/hiveot/hub/pkg/gateway"
	"net"
	"os"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
)

// ServeDiscovery publishes a HiveOT service for DNS-SD discovery on the local domain.
// Intended only for services that listen on their own TCP port.
//
// See also 'DiscoverService' for discovery of this published service.
//
// If no address is provided, this uses the external facing address (eg the one with access to the internet)
//
// Intended for Hub Services (eg the gateway) to announce a TCP and Websocket reachable service and how they can be reached on the
// local domain.
//
// DNS-SD will publish this as <instanceID>._hiveot._tcp
//
//	serviceName is the name of the hiveot service being published, eg gateway
//	serviceType is the name of the protocol service, "" is default 'hiveot'
//	address is the listener address or "" for the outbound IP (eg not localhost)
//	tcpPort is the listening port for TCP connections
//	wssPort is the listener port for WebSocket connections or 0 if n/a
//	connectionPath is the websocket connection path or "" if n/a
//
// Returns the discovery service instance. Use Shutdown() when done.
func ServeDiscovery(serviceName string, serviceType string,
	address string, tcpPort int, wssPort int, connectionPath string) (*zeroconf.Server, error) {

	var ips = make([]string, 0)
	textRecord := make([]string, 0)
	hostName, _ := os.Hostname()
	instanceID := serviceName + "-" + hostName
	if serviceName == "" {
		return nil, fmt.Errorf("missing serviceName")
	}

	logrus.Infof("instanceID=%s, serviceType=%s, address=%s, tcpPort=%d, wssPort=%d, connectionPath=%s",
		instanceID, serviceType, address, tcpPort, wssPort, connectionPath)
	protocolType := "_" + serviceType + "._tcp"
	if serviceType == "" {
		protocolType = gateway.HIVEOT_DNSSD_TYPE
	}

	// only the local domain is supported
	domain := "local"
	hostname, _ := os.Hostname()

	if address == "" {
		ip := GetOutboundIP("")
		ips = append(ips, ip.String())
	} else {
		ip := net.ParseIP(address)
		if ip == nil {
			err := fmt.Errorf("provided address '%s' is not a valid IP address", address)
			logrus.Error(err)
			return nil, err
		}
		ips = append(ips, ip.String())
	}
	// ignore lo as this faces the local network
	ifaces, err := GetInterfaces(ips[0])
	if err != nil || len(ifaces) == 0 {
		logrus.Warningf("Address %s does not appear on any interface. Continuing anyways", ips[0])
	}
	// add a text record with key=value pairs for websocket support
	textRecord = append(textRecord, "path="+connectionPath)
	if wssPort != 0 {
		textRecord = append(textRecord, fmt.Sprintf("wss=%d", wssPort))
	}
	server, err := zeroconf.RegisterProxy(
		instanceID, protocolType, domain, tcpPort, hostname, ips, textRecord, ifaces)
	if err != nil {
		logrus.Errorf("Failed to start the zeroconf server: %s", err)
	}
	return server, err
}
