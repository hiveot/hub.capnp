package hubnet

import (
	"fmt"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetOutboundInterface Get preferred outbound network interface of this machine
// Credits: https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
// and https://qiita.com/shaching/items/4c2ee8fd2914cce8687c
func GetOutboundInterface(address string) (interfaceName string, macAddress string, ipAddr net.IP) {
	if address == "" {
		address = "1.1.1.1"
	}

	// This dial command doesn't actually create a connection
	conn, err := net.Dial("udp", address+":9999")
	if err != nil {
		logrus.Errorf("GetOutboundInterface for address '%s': %s", address, err)
		return "", "", nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ipAddr = localAddr.IP

	// find the first interface for this address
	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {

		if addrs, err := interf.Addrs(); err == nil {
			for index, addr := range addrs {
				logrus.Debug("[", index, "]", interf.Name, ">", addr)

				// only interested in the name with current IP address
				if strings.Contains(addr.String(), ipAddr.String()) {
					logrus.Debug("GetOutboundInterface: Use name : ", interf.Name)
					interfaceName = interf.Name
					macAddress = fmt.Sprint(interf.HardwareAddr)
					break
				}
			}
		}
	}
	return
}
