// Package internal with launching of the mosquitto broker
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LaunchMosquitto launches mosquitto with the given configuration file. This attaches stderr and stdout
// to the current process.
// This signals the 'done' channel when the mosquitto process has ended
//  returns with the command or error. Use cmd.Process.Kill to terminate.
func LaunchMosquitto(configFile string, done chan bool) (*exec.Cmd, error) {
	logrus.Infof("Starting mosquitto broker")
	var isRunning bool
	var mu sync.Mutex

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
		t1 := time.Now()
		mu.Lock()
		isRunning = true
		mu.Unlock()
		// logrus.Infof("Mosquitto cmd.Wait started")
		err = cmd.Wait()
		done <- true
		duration := time.Since(t1)
		logrus.Infof("Mosquitto has ended after %.3f seconds. err=%v", duration.Seconds(), err)
		mu.Lock()
		isRunning = false
		mu.Unlock()
	}()
	// Give mosquitto some time to start, if starting failed due to error we might pick it up
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	if !isRunning && err == nil {
		err = fmt.Errorf("Mosquitto terminated immediately; check the template and the logs for info")
	}
	mu.Unlock()

	return cmd, err
}
