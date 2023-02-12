// Package watcher that handles file renames
package watcher

import (
	"context"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// debounce timer that invokes the callback after no more changes are received
const watcherDebounceDelay = 50 // msec

// WatchFile is a resilient file watcher that handles file renames
// Special features:
//  1. This debounces multiple quick changes before invoking the callback
//  2. After the callback, resubscribe to the file to handle file renames that change the file inode
//     path to watch
//     handler to invoke on change
//
// This returns the fsnotify watcher. Close it when done.
func WatchFile(ctx context.Context, path string,
	handler func(ctx context.Context) error) (*fsnotify.Watcher, error) {
	_ = ctx
	watcher, _ := fsnotify.NewWatcher()
	// The callback timer debounces multiple changes to the config file
	callbackTimer := time.AfterFunc(0, func() {
		logrus.Infof("WatchFile.Watch: trigger, invoking callback...")
		_ = handler(ctx)

		// file renames change the inode of the filename, resubscribe
		_ = watcher.Remove(path)
		err := watcher.Add(path)
		if err != nil {
			logrus.Errorf("failed adding file to watch '%s': %s", path, err)
		}
	})
	callbackTimer.Stop() // don't start yet

	err := watcher.Add(path)
	if err != nil {
		logrus.Infof("unable to watch for changes: %s", err)
		return watcher, err
	}
	logrus.Infof("WatchFile added path: %s", path)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logrus.Infof("No more events. Ending watch for file='%s'. event=%s", path, event)
					callbackTimer.Stop()
					return
				}
				// don't really care what the change it, 50msec after the last event the file will reload
				logrus.Infof("Event: '%s'. Modified file: %s", event, event.Name)
				callbackTimer.Reset(time.Millisecond * watcherDebounceDelay)
			case err2, ok := <-watcher.Errors:
				if !ok && err2 != nil {
					logrus.Errorf("Unexpected error: %s", err2)
					return
				}
				// end of watcher.
				//logrus.Errorf("Error: %s", err2)
			}
		}
	}()
	return watcher, nil
}
