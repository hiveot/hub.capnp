package service

import (
	"context"

	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/directory"
)

// UpdateDirectory is a provides the capability to update the directory
// This implements the IUpdateDirectory API
type UpdateDirectory struct {
	// bucket that holds the TD documents
	bucket bucketstore.IBucket
}

func (svc *UpdateDirectory) RemoveTD(_ context.Context, thingID string) error {
	err := svc.bucket.Delete(thingID)
	return err
}

func (svc *UpdateDirectory) UpdateTD(_ context.Context, id string, td []byte) error {
	err := svc.bucket.Set(id, td)
	return err
}

func (svc *UpdateDirectory) Release() {
	_ = svc.bucket.Close()
}

// NewUpdateDirectory returns the capability to update the directory
// bucket with the TD documents. Will be closed when done.
func NewUpdateDirectory(bucket bucketstore.IBucket) directory.IUpdateDirectory {
	svc := &UpdateDirectory{bucket: bucket}
	return svc
}
