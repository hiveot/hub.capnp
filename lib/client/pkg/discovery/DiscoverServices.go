// Package discovery with client for service discovery
package discovery

import (
	"fmt"
	"strings"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
)

// DiscoverServices searches for WoST services with the given name and returns all its instances.
// This is a wrapper around various means of discovering services and supports the discovery of multiple
// instances of the same service (name). The serviceName must contain the simple name of the WoST service.
// For example, use 'idprov' for the provisioning service which DNS-SD will publish as _idprov._tcp.
//
//  serviceName is the name of the service to discover
//  waitSec is the time to wait for the result
// Returns the first instance address, port and discovery parameters, plus records of additional discoveries,
// or an error if nothing is found
func DiscoverServices(serviceName string, waitSec int) (
	address string, port uint, params map[string]string,
	records []*zeroconf.ServiceEntry, err error) {
	params = make(map[string]string)

	serviceType := fmt.Sprintf("_%s._tcp", serviceName)
	records, err = DnsSDScan(serviceType, waitSec)
	if err != nil {
		return "", 0, nil, nil, err
	}
	if len(records) == 0 {
		err = fmt.Errorf("DiscoverService: No service of name '%s' (serviceType=%s) found after %d seconds", serviceName, serviceType, waitSec)
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
			logrus.Infof("DiscoverService: Ignoring non key-value '%s' in TXT record", txtRecord)
		} else {
			params[kv[0]] = kv[1]
		}
	}
	return address, uint(rec0.Port), params, records, nil
}
