package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/history"
	"github.com/hiveot/hub/pkg/history/config"
	"github.com/hiveot/hub/pkg/pubsub"
)

const PropertiesBucketName = "properties"

// HistoryService provides storage for action and event history using the bucket store
// Each Thing has a bucket with events and actions.
// This implements the IHistoryService interface
type HistoryService struct {

	// The history service bucket store with a bucket for each Thing
	bucketStore bucketstore.IBucketStore
	// Storage of the latest properties of a thing
	propsStore *LastPropertiesStore
	// handling of retention of pubsub events
	retentionMgr *ManageRetention
	// Instance ID of this service
	serviceID string
	// the pubsub service to subscribe to event
	servicePubSub pubsub.IServicePubSub
	// optional handling of pubsub events. nil if not used
	subEventHandler *PubSubEventHandler
}

// CapAddHistory provides the capability to add to the history of any Thing.
// This capability should only be provided to trusted services that capture events from multiple sources
// and can verify their authenticity.
func (svc *HistoryService) CapAddHistory(
	_ context.Context, clientID string, ignoreRetention bool) (history.IAddHistory, error) {

	logrus.Infof("clientID=%s", clientID)

	var retentionMgr *ManageRetention
	if !ignoreRetention {
		retentionMgr = svc.retentionMgr
	}

	historyUpdater := NewAddHistory(
		clientID, svc.bucketStore, retentionMgr, svc.propsStore.HandleAddValue)
	return historyUpdater, nil
}

// CapManageRetention returns the capability to manage the retention of events
func (svc *HistoryService) CapManageRetention(
	_ context.Context, clientID string) (history.IManageRetention, error) {

	logrus.Infof("clientID=%s", clientID)
	evRet := svc.retentionMgr
	_ = clientID
	return evRet, nil
}

// CapReadHistory provides the capability to read history
func (svc *HistoryService) CapReadHistory(
	_ context.Context, clientID, publisherID, thingID string) (history.IReadHistory, error) {

	logrus.Infof("clientID=%s", clientID)
	thingAddr := publisherID + "/" + thingID
	bucket := svc.bucketStore.GetBucket(thingAddr)
	readHistory := NewReadHistory(clientID, publisherID, thingID, bucket, svc.propsStore.GetProperties)
	return readHistory, nil
}

// Start using the history service
// This will open the store and panic if the store cannot be opened.
func (svc *HistoryService) Start() (err error) {
	logrus.Infof("")
	err = svc.bucketStore.Open()
	if err != nil {
		logrus.Panic("can't open histroy store")
	}
	propsbucket := svc.bucketStore.GetBucket(PropertiesBucketName)
	svc.propsStore = NewPropertiesStore(propsbucket)

	err = svc.retentionMgr.Start()

	// subscribe to events to add history
	if err == nil && svc.servicePubSub != nil {
		capAddEvent := NewAddHistory(svc.serviceID, svc.bucketStore, svc.retentionMgr, svc.propsStore.HandleAddValue)
		svc.subEventHandler = NewSubEventHandler(svc.servicePubSub, capAddEvent)
		err = svc.subEventHandler.Start()
	}

	return err
}

// Stop using the history service and release resources
func (svc *HistoryService) Stop() error {
	logrus.Infof("")
	err := svc.propsStore.SaveChanges()
	if err != nil {
		logrus.Error(err)
	}
	svc.retentionMgr.Stop()
	if svc.subEventHandler != nil {
		svc.subEventHandler.Stop()
	}
	err = svc.bucketStore.Close()
	if err != nil {
		logrus.Error(err)
	}
	return err
}

// NewHistoryService creates a new instance for the history service using the given
// storage bucket.
//
//	serviceID is the thingID of the service
//	store contains the bucket store to use. This will be opened on Start() and closed on Stop()
//	sub pubsub client to store events. nil to not subscribe to events. Will be released on Stop().
func NewHistoryService(
	config *config.HistoryConfig, store bucketstore.IBucketStore, sub pubsub.IServicePubSub) *HistoryService {

	if config.ServiceID == "" {
		config.ServiceID = history.ServiceName
	}
	retentionMgr := NewManageRetention(config.Retention)
	svc := &HistoryService{
		bucketStore:   store,
		propsStore:    nil,
		serviceID:     config.ServiceID,
		retentionMgr:  retentionMgr,
		servicePubSub: sub,
	}
	return svc
}
