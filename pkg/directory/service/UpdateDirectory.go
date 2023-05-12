package service

import (
	"context"
	"encoding/json"
	"github.com/hiveot/hub/api/go/hubapi"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/api/go/vocab"
	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/directory"
)

// UpdateDirectory is a provides the capability to update the directory
// This implements the IUpdateDirectory API
//
//	Bucket keys are made of gatewayID+"/"+thingID
//	Bucket values are ThingValue objects
type UpdateDirectory struct {
	// The client that is updating the directory
	clientID string
	// bucket that holds the TD documents
	bucket bucketstore.IBucket
}

func (svc *UpdateDirectory) RemoveTD(_ context.Context, publisherID, thingID string) error {
	logrus.Infof("clientID=%s, thingID=%s", svc.clientID, thingID)
	thingAddr := publisherID + "/" + thingID
	err := svc.bucket.Delete(thingAddr)
	return err
}

func (svc *UpdateDirectory) UpdateTD(_ context.Context, publisherID, thingID string, td []byte) error {
	//logrus.Infof("clientID=%s, thingID=%s", svc.clientID, thingID)

	bucketValue := &thing.ThingValue{
		PublisherID: publisherID,
		ThingID:     thingID,
		ID:          hubapi.EventNameTD,
		Data:        td,
		Created:     time.Now().Format(vocab.ISO8601Format),
	}
	bucketData, _ := json.Marshal(bucketValue)
	thingAddr := publisherID + "/" + thingID
	err := svc.bucket.Set(thingAddr, bucketData)
	return err
}

func (svc *UpdateDirectory) Release() {
	_ = svc.bucket.Close()
}

// NewUpdateDirectory returns the capability to update the directory
// bucket with the TD documents. Will be closed when done.
func NewUpdateDirectory(clientID string, bucket bucketstore.IBucket) directory.IUpdateDirectory {
	svc := &UpdateDirectory{
		clientID: clientID,
		bucket:   bucket,
	}
	return svc
}
