package watcher

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// debounce timer that invokes the callback after no more changes are received
const watcherDebounceDelay = 50 // msec

// WatchFile is a resilient file watcher that handles file renames
// Special features:
// 1. This debounces multiple quick changes before invoking the callback
// 2. After the callback, resubscribe to the file to handle file renames that change the file inode
//	clientID for logging of who is doing the watching.
//  path to watch
//  handler to invoke on change
// This returns the fsnotify watcher. Close it when done.
func WatchFile(path string, handler func() error, clientID string) (*fsnotify.Watcher, error) {
	watcher, _ := fsnotify.NewWatcher()
	// The callback timer debounces multiple changes to the config file
	callbackTimer := time.AfterFunc(0, func() {
		logrus.Infof("WatchFile.Watch: trigger, invoking callback for clientID='%s'", clientID)
		handler()

		// file renames change the inode of the filename, resubscribe
		watcher.Remove(path)
		watcher.Add(path)
	})
	callbackTimer.Stop() // don't start yet

	err := watcher.Add(path)
	if err != nil {
		logrus.Errorf("WatchFile.Watch: clientID='%s', unable to watch for changes: %s", clientID, err)
		return watcher, err
	}
	logrus.Infof("WatchFile.Watch: clientID='%s', added path: %s", clientID, path)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logrus.Warningf("WatchFile: clientID='%s'. No more events. Ending watch for file='%s'. event=%s",
						clientID, path, event)
					return
				}
				// don't really care what the change it, 50msec after the last event the file will reload
				logrus.Infof("WatchFile: clientID='%s', event: '%s'. Modified file: %s",
					clientID, event, event.Name)
				callbackTimer.Reset(time.Millisecond * watcherDebounceDelay)
			case err, ok := <-watcher.Errors:
				if !ok {
					logrus.Errorf("WatchFile: clientID='%s', Unexpected error: %s", clientID, err)
					return
				}
				logrus.Errorf("WatchFile: clientID='%s', Error: %s", clientID, err)
			}
		}
	}()
	return watcher, nil
}
