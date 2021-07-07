// package main for both the protocol binding and the mosquitto auth plugin
package main

// #cgo CFLAGS: -g  -fPIC -I/usr/local/include -I./
// #cgo LDFLAGS: -L.
import "C"
import (
	mosquittopb "github.com/wostzone/hub/core/mosquitto-pb"
	"github.com/wostzone/hub/core/mosquitto-pb/mosqplug"
)

var pluginConfig = &mosquittopb.PluginConfig{}

//export AuthPluginInit
func AuthPluginInit(keys []string, values []string, authOptsNum int) {
	mosqplug.AuthPluginInit(keys, values, authOptsNum)
}

//export AuthUnpwdCheck
func AuthUnpwdCheck(username string, password string, clientID string, clientIP string) uint8 {
	return mosqplug.AuthUnpwdCheck(username, password, clientID, clientIP)
}

//export AuthAclCheck
// certAuth is 1 for cert authenticated clients, 0 for pw authenticated clients
func AuthAclCheck(clientid, username, topic string, acc int, certAuth int) uint8 {
	return mosqplug.AuthAclCheck(clientid, username, topic, acc, certAuth != 0)
}

//export AuthPluginCleanup
func AuthPluginCleanup() {
	mosqplug.AuthPluginCleanup()
}

func main() {}

// Main entry to WoST protocol adapter for managing Mosquitto
// This setup the configuration from file and commandline parameters and launches the service
// func main() {
// 	svc := mosquittopb.NewMosquittoManager()
// 	hubConfig, err := hubconfig.LoadCommandlineConfig("", mosquittopb.PluginID, &svc.Config)
// 	if err != nil {
// 		logrus.Errorf("ERROR: Start aborted due to error")
// 		os.Exit(1)
// 	}

// 	err = svc.Start(hubConfig)
// 	if err != nil {
// 		logrus.Errorf("Logger: Failed to start: %s", err)
// 		os.Exit(1)
// 	}
// 	hubclient.WaitForSignal()
// 	svc.Stop()
// 	os.Exit(0)
// }
