package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub.go/pkg/vocab"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/hiveot/hub/pkg/pubsub"
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

func (svc *UpdateDirectory) RemoveTD(_ context.Context, thingAddr string) error {
	err := svc.bucket.Delete(thingAddr)
	return err
}

func (svc *UpdateDirectory) UpdateTD(_ context.Context, thingAddr string, td []byte) error {

	bucketValue := &thing.ThingValue{
		ThingAddr: thingAddr,
		Name:      pubsub.MessageTypeTD,
		ValueJSON: td,
		Created:   time.Now().Format(vocab.ISO8601Format),
	}
	bucketData, _ := json.Marshal(bucketValue)
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
