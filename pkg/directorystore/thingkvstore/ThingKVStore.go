package thingkvstore

import (
	"context"

	"github.com/hiveot/hub/internal/kvstore"
)

// ThingKVStoreServer is a thing wrapper around the internal KVStore
type ThingKVStoreServer struct {
	store *kvstore.KVStore
}

// ListTDs returns an array of TD documents in JSON text
//  thingIDs optionally restricts the result to the given IDs
func (srv *ThingKVStoreServer) ListTDs(_ context.Context, limit int, offset int, thingIDs []string) ([]string, error) {
	res := make([]string, 0)
	docs, err := srv.store.List(limit, offset, thingIDs)
	if err == nil {
		for _, doc := range docs {
			res = append(res, doc)
		}
	}
	return res, err
}

// QueryTDs returns an array of TD documents that match the jsonPath query
//  thingIDs optionally restricts the result to the given IDs
func (srv *ThingKVStoreServer) QueryTDs(_ context.Context, jsonPathQuery string, limit int, offset int, thingIDs []string) ([]string, error) {

	resp, err := srv.store.Query(jsonPathQuery, limit, offset, thingIDs)
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

// GetTD returns the TD document with the given ID
func (srv *ThingKVStoreServer) GetTD(_ context.Context, thingID string) (string, error) {
	return srv.store.Read(thingID)
	//resp, err := srv.store.Read(thingID)
	//if err == nil {
	//	err = json.Unmarshal([]byte(resp), &res)
	//}
	//return res, err
}

func (srv *ThingKVStoreServer) RemoveTD(_ context.Context, thingID string) error {
	srv.store.Remove(thingID)
	return nil
}

func (srv *ThingKVStoreServer) UpdateTD(_ context.Context, id string, td string) error {
	err := srv.store.Write(id, td)
	return err
}

// NewThingKVStoreServer creates a service to access TDs in the state store
//  storeName is the name of the TD state store. Default is 'DefaultTDStoreName'
func NewThingKVStoreServer(thingStorePath string) (*ThingKVStoreServer, error) {

	kvStore, err := kvstore.NewKVStore(thingStorePath)
	srv := &ThingKVStoreServer{
		store: kvStore,
	}
	return srv, err
}
