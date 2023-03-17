// Package hubclient with client for Hub gateway service discovery
package hubclient

import (
	"fmt"
	"github.com/grandcat/zeroconf"
	"github.com/hiveot/hub/pkg/gateway"
	"github.com/hiveot/hub/pkg/resolver"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

// DiscoverService searches for services with the given type and returns all its instances.
// This is a wrapper around various means of discovering services and supports the discovery of multiple
// instances of the same service (name). The serviceName must contain the simple name of the Hub service.
// For example, use 'idprov' for the provisioning service which DNS-SD will publish as _idprov._tcp.
//
//	serviceType is the type of service to discover without the "_", eg "hiveot" in "_hiveot._tcp"
//	waitSec is the time to wait for the result
//
// Returns the first instance address, port and discovery parameters, plus records of additional discoveries,
// or an error if nothing is found
func DiscoverService(serviceType string, waitSec int) (
	address string, port int, params map[string]string,
	records []*zeroconf.ServiceEntry, err error) {
	params = make(map[string]string)

	serviceProtocol := "_" + serviceType + "._tcp"
	if serviceType == "" {
		serviceProtocol = gateway.HIVEOT_DNSSD_TYPE
	}
	records, err = DnsSDScan(serviceProtocol, waitSec)
	if err != nil {
		return "", 0, nil, nil, err
	}
	if len(records) == 0 {
		err = fmt.Errorf("no service of type '%s' found after %d seconds", serviceProtocol, waitSec)
		return "", 0, nil, nil, err
	}
	rec0 := records[0]

	// determine the address string
	// use the local IP if provided
	if len(rec0.AddrIPv4) > 0 {
		address = rec0.AddrIPv4[0].String()
	} else if len(rec0.AddrIPv6) > 0 {
		address = rec0.AddrIPv6[0].String()
	} else {
		// fall back to use host.domainname
		address = rec0.HostName
	}

	// reconstruct key-value parameters from TXT record
	for _, txtRecord := range rec0.Text {
		kv := strings.Split(txtRecord, "=")
		if len(kv) != 2 {
			logrus.Infof("Ignoring non key-value '%s' in TXT record", txtRecord)
		} else {
			params[kv[0]] = kv[1]
		}
	}
	return address, rec0.Port, params, records, nil
}

// LocateHub determines the hiveot full URL for the given transport, "unix", "tcp", "wss" or "" for any
// This first checks if a UDS socket exists on the default resolver-service path.
// Secondly, perform a DNS-SD search. If multiple results are returned then preference goes to "tcp" transport.
func LocateHub(transport string, searchTime int) (fullURL string) {

	// prefer the default resolver using unix domain sockets
	if transport == "" || transport == "unix" {
		if _, err := os.Stat(resolver.DefaultResolverPath); err == nil {
			//fullURL = "unix://" + resolver.DefaultResolverPath
			//return fullURL
		} else {
			// not found, continue the search
		}
	}
	// discover the service and determine the best matching record
	// yes, this seems like a bit of a pain
	// default is the hiveot service
	addr, port, params, records, err := DiscoverService("hiveot", searchTime)
	logrus.Infof("Found %d records. Using the first record.", len(records))
	if err != nil {
		// failed, nothing to be found
		return ""
	}
	// default value
	if transport == "" || transport == "tcp" {
		fullURL = fmt.Sprintf("%s:%d", addr, port)
	} else if transport == "wss" {
		wssPort, found := params["wss"]
		if found {
			fullURL = fmt.Sprintf("%s:%d%s", addr, wssPort, params["path"])
		}
	}
	return
}
