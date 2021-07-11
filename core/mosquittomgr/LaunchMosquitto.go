// Package mosquittomgr with launching of the mosquitto broker
package mosquittomgr

import (
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

// launch mosquitto with the given configuration file. This attaches stderr and stdout
// to the current process.
//  returns with the command or error. Use cmd.Process.Kill to terminate.
func LaunchMosquitto(configFile string) (*exec.Cmd, error) {

	logrus.Infof("--- Starting mosquitto broker ---")

	// mosquitto must be in the path
	cmd := exec.Command("mosquitto", "-c", configFile)
	// Capture stderr in case of startup failure
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	go func() {
		cmd.Wait()
		logrus.Infof("--- Mosquitto has ended ---")
	}()
	// Give mosquitto some time to start
	time.Sleep(10 * time.Millisecond)

	return cmd, err
}
