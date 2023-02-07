package service

import (
	"context"
	"encoding/json"

	"github.com/hiveot/hub/lib/thing"
	"github.com/hiveot/hub/pkg/bucketstore"
	"github.com/hiveot/hub/pkg/directory"
	"github.com/sirupsen/logrus"
)

// ReadDirectory is a provides the capability to read and iterate the directory
// This implements the IReadDirectory API
type ReadDirectory struct {
	// the client that is reading the directory
	clientID string
	// read bucket that holds the TD documents
	bucket bucketstore.IBucket
}

// GetTD returns the TD document for the given Thing ID in JSON format
func (svc *ReadDirectory) GetTD(_ context.Context, publisherID, thingID string) (tdValue *thing.ThingValue, err error) {
	logrus.Infof("clientID=%s, thingID=%s", svc.clientID, thingID)
	// bucket keys are made of the gatewayID / thingID
	thingAddr := publisherID + "/" + thingID
	raw, err := svc.bucket.Get(thingAddr)
	if raw != nil {
		err = json.Unmarshal(raw, &tdValue)
	}
	return tdValue, err
}

// Cursor returns an iterator for ThingValues containing a TD document
func (svc *ReadDirectory) Cursor(_ context.Context) (cursor directory.IDirectoryCursor) {
	logrus.Infof("clientID=%s", svc.clientID)
	dirCursor := NewDirectoryCursor(svc.bucket.Cursor())
	return dirCursor

}

//// ListTDs returns an array of TD documents in JSON text
//func (srv *DirectoryKVStoreServer) ListTDs(_ context.Context, limit int, offset int) ([]string, error) {
//	res := make([]string, 0)
//	docs, err := srv.store.List(srv.defaultBucket, limit, offset, nil)
//	if err == nil {
//		for _, doc := range docs {
//			res = append(res, doc)
//		}
//	}
//	return res, err
//}

// ListTDcb provides a callback with an array of TD documents in JSON text
//func (srv *DirectoryKVStoreServer) ListTDcb(
//	ctx context.Context, handler func(td string, isLast bool) error) error {
//	_ = ctx
//	batch := make([]string, 0)
//	docs, err := srv.store.List(srv.defaultBucket, 0, 0, nil)
//	if err == nil {
//		// convert map to array
//		for _, doc := range docs {
//			batch = append(batch, doc)
//		}
//		// for testing, callback one at a time
//		//err = handler(batch, true)
//		for i, tddoc := range batch {
//			docList := []string{tddoc}
//			isLast := i == len(batch)-1
//			err = handler(docList, isLast)
//		}
//	}
//	return err
//}

// QueryTDs returns an array of TD documents that match the jsonPath query
//  thingIDs optionally restricts the result to the given IDs
//func (srv *DirectoryKVStoreServer) QueryTDs(_ context.Context, jsonPathQuery string, limit int, offset int) ([]string, error) {
//
//	resp, err := srv.store.Query(jsonPathQuery, limit, offset, nil)
//	return resp, err
//	//res := make([]string, 0)
//	//if err == nil {
//	//	for _, docText := range resp {
//	//		var td thing.ThingDescription
//	//		err = json.Unmarshal([]byte(docText), &td)
//	//		res.Things = append(res.Things, &td)
//	//	}
//	//}
//	//return res, err
//}

// QueryTDs returns the TD's filtered using JSONpath on the TD content
// See 'docs/query-tds.md' for examples
// disabled as this is not used
//QueryTDs(ctx context.Context, jsonPath string, limit int, offset int) (tds []string, err error)

// Release this capability and allocated resources after its use
func (svc *ReadDirectory) Release() {
	// logrus.Infof("Released")
	err := svc.bucket.Close()
	_ = err
}

// NewReadDirectory returns the capability to read the directory
// bucket with the TD documents. Will be closed when done.
func NewReadDirectory(clientID string, bucket bucketstore.IBucket) directory.IReadDirectory {
	// logrus.Infof("NewReadDirectory for bucket: ", bucket.ID())
	svc := &ReadDirectory{
		clientID: clientID,
		bucket:   bucket,
	}
	return svc
}
