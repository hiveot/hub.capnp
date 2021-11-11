package hubnet

import (
	"net"

	"github.com/sirupsen/logrus"
)

// GetOutboundIP returns the default outbound IP address to reach the given hostname.
// Use a local hostname if a subnet other than the default one should be used.
// Use "" for the default route address
//  destination to reach or "" to use 1.1.1.1 (no connection will be established)
func GetOutboundIP(destination string) net.IP {
	if destination == "" {
		destination = "1.1.1.1"
	}
	// This dial command doesn't actually create a connection
	conn, err := net.Dial("udp", destination+":80")
	if err != nil {
		logrus.Errorf("GetIPAddr: %s", err)
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}
