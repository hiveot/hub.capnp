package idprovclient

import (
	"fmt"

	"github.com/wostzone/hub/lib/client/pkg/discovery"
)

const IdprovServiceName = "idprov"

// DiscoverProvisioningServer provides the URL of the first instance of a provisioning server
// on the local network.
//
//  serviceName of the IDProv server. Use "" for default ('idprov')
//  waitSec is the time to wait for the result
// Returns the URL of the IDProv directory endpoint on the local network
func DiscoverProvisioningServer(serviceName string, waitSec int) (string, error) {
	var url string
	addr, port, params, records, err := discovery.DiscoverServices(serviceName, waitSec)
	_ = records
	if err != nil {
		err = fmt.Errorf("no provisioning server found after %d seconds", waitSec)
		return "", err
	}
	path := params["path"]
	if path == "" {
		err = fmt.Errorf("provisioning service discovery is missing the endpoint path")
		return "", err
	}

	// fall back to use host.domainname
	url = fmt.Sprintf("https://%s:%d%s", addr, port, path)

	return url, nil
}
