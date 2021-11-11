// Package discovery with client for service discovery
package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
)

// DnsSdScan scans zeroconf publications on local domain
// The zeroconf library does not support browsing of all services, but a workaround is
// to search the service types with "_services._dns-sd._udp" then query each of the service types.
//  serviceType to look for, or "" to discover all service types (not all services)
func DnsSDScan(sdType string, waitSec int) ([]*zeroconf.ServiceEntry, error) {
	sdDomain := "local."
	mu := &sync.Mutex{}

	if sdType == "" {
		// https://github.com/grandcat/zeroconf/pull/15
		sdType = "_services._dns-sd._udp"
	}
	records := make([]*zeroconf.ServiceEntry, 0)
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		logrus.Errorf("DnsSDScan: Failed to create resolver instance: %s", err)
		return nil, err
	}

	// 'records' channel captures the result
	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			rec := entry.ServiceRecord
			logrus.Infof("DnsSDScan: Found service instance '%s' of type '%s', domain '%s'. ip4:port=%s:%d",
				rec.Instance, rec.ServiceName(), rec.Domain, entry.AddrIPv4, entry.Port)
			mu.Lock()
			records = append(records, entry)
			mu.Unlock()
		}
		logrus.Infof("DnsSDScan: No more entries.")
	}(entries)

	duration := time.Second * time.Duration(waitSec)
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	err = resolver.Browse(ctx, sdType, sdDomain, entries)
	if err != nil {
		logrus.Fatalf("DnsSDScan: Failed to browse: %s", err.Error())
	}
	<-ctx.Done()
	mu.Lock()
	results := records
	mu.Unlock()

	return results, err
}
