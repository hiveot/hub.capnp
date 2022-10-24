package directorykvstore

import (
	"context"
	"encoding/json"

	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub.go/pkg/vocab"
	"github.com/hiveot/hub/internal/kvstore"
	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryKVStoreServer is a thing wrapper around the internal KVStore
// This implements the IDirectory, IReadDirectory and IUpdateDirectory interfaces
type DirectoryKVStoreServer struct {
	store *kvstore.KVStore
}

// Create a new Thing TD document describing this service
func (srv *DirectoryKVStoreServer) createServiceTD() *thing.ThingDescription {
	thingID := thing.CreateThingID("", directory.ServiceName, vocab.DeviceTypeService)
	title := "Directory KV Store Server"
	deviceType := vocab.DeviceTypeService
	td := thing.CreateTD(thingID, title, deviceType)

	return td
}

// CapReadDirectory provides the service to read the directory
func (srv *DirectoryKVStoreServer) CapReadDirectory(ctx context.Context) directory.IReadDirectory {
	return srv
}

// CapUpdateDirectory provides the service to update the directory
func (srv *DirectoryKVStoreServer) CapUpdateDirectory(ctx context.Context) directory.IUpdateDirectory {
	return srv
}

// GetTD returns the TD document with the given ID
func (srv *DirectoryKVStoreServer) GetTD(_ context.Context, thingID string) (string, error) {
	return srv.store.Read(thingID)
	//resp, err := srv.store.Read(thingID)
	//if err == nil {
	//	err = json.Unmarshal([]byte(resp), &res)
	//}
	//return res, err
}

// ListTDs returns an array of TD documents in JSON text
func (srv *DirectoryKVStoreServer) ListTDs(_ context.Context, limit int, offset int) ([]string, error) {
	res := make([]string, 0)
	docs, err := srv.store.List(limit, offset, nil)
	if err == nil {
		for _, doc := range docs {
			res = append(res, doc)
		}
	}
	return res, err
}

// ListTDcb provides a callback with an array of TD documents in JSON text
func (srv *DirectoryKVStoreServer) ListTDcb(
	ctx context.Context, handler func(batch []string, isLast bool) error) error {
	_ = ctx
	batch := make([]string, 0)
	docs, err := srv.store.List(0, 0, nil)
	if err == nil {
		// convert map to array
		for _, doc := range docs {
			batch = append(batch, doc)
		}
		// for testing, callback one at a time
		//err = handler(batch, true)
		for i, tddoc := range batch {
			docList := []string{tddoc}
			isLast := i == len(batch)-1
			err = handler(docList, isLast)
		}
	}
	return err
}

// QueryTDs returns an array of TD documents that match the jsonPath query
//  thingIDs optionally restricts the result to the given IDs
func (srv *DirectoryKVStoreServer) QueryTDs(_ context.Context, jsonPathQuery string, limit int, offset int) ([]string, error) {

	resp, err := srv.store.Query(jsonPathQuery, limit, offset, nil)
	return resp, err
	//res := make([]string, 0)
	//if err == nil {
	//	for _, docText := range resp {
	//		var td thing.ThingDescription
	//		err = json.Unmarshal([]byte(docText), &td)
	//		res.Things = append(res.Things, &td)
	//	}
	//}
	//return res, err
}

func (srv *DirectoryKVStoreServer) RemoveTD(_ context.Context, thingID string) error {
	srv.store.Remove(thingID)
	return nil
}

func (srv *DirectoryKVStoreServer) UpdateTD(_ context.Context, id string, td string) error {
	err := srv.store.Write(id, td)
	return err
}

// Start creates the store and updates the service own TD
func (srv *DirectoryKVStoreServer) Start(ctx context.Context) error {
	err := srv.store.Start()
	if err == nil {
		myTD := srv.createServiceTD()
		myTDJson, _ := json.Marshal(myTD)
		err = srv.UpdateTD(ctx, myTD.ID, string(myTDJson))
	}
	return err
}

// Stop the storage server and flush changes to disk
func (srv *DirectoryKVStoreServer) Stop() {
	_ = srv.store.Stop()
}

// NewDirectoryKVStoreServer creates a service to access TDs in the state store
//  thingStorePath is the file holding the directory data.
func NewDirectoryKVStoreServer(ctx context.Context, thingStorePath string) (*DirectoryKVStoreServer, error) {

	kvStore, err := kvstore.NewKVStore(thingStorePath)
	srv := &DirectoryKVStoreServer{
		store: kvStore,
	}
	if err == nil {
		err = srv.Start(ctx)
	}
	return srv, err
}
