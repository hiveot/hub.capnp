package directorykvstore

import (
	"context"

	"github.com/hiveot/hub/internal/kvstore"
	"github.com/hiveot/hub/pkg/directory"
)

// DirectoryKVStoreServer is a thing wrapper around the internal KVStore
// This implements the IDirectory, IReadDirectory and IUpdateDirectory interfaces
type DirectoryKVStoreServer struct {
	store *kvstore.KVStore
}

// CapReadDirectory provides the service to read the directory
func (srv *DirectoryKVStoreServer) CapReadDirectory() directory.IReadDirectory {
	return srv
}

// CapUpdateDirectory provides the service to update the directory
func (srv *DirectoryKVStoreServer) CapUpdateDirectory() directory.IUpdateDirectory {
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
//  thingIDs optionally restricts the result to the given IDs
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

// NewDirectoryKVStoreServer creates a service to access TDs in the state store
//  storeName is the name of the TD state store. Default is 'DefaultTDStoreName'
func NewDirectoryKVStoreServer(thingStorePath string) (*DirectoryKVStoreServer, error) {

	kvStore, err := kvstore.NewKVStore(thingStorePath)
	srv := &DirectoryKVStoreServer{
		store: kvStore,
	}
	return srv, err
}
