package thingkvstore

import (
	"context"
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/wostzone/hub/internal/kvstore"

	"github.com/wostzone/wost.grpc/go/svc"
	"github.com/wostzone/wost.grpc/go/thing"
)

// ThingKVStoreServer implements the svc.ThingStore interface using the internal KVStore
// TDs are kept in memory for quick response and search indexing
type ThingKVStoreServer struct {
	svc.UnimplementedThingStoreServer
	store *kvstore.KVStore
}

func (srv *ThingKVStoreServer) List(_ context.Context, args *svc.ListTD_Args) (*thing.ThingDescriptionList, error) {
	res := &thing.ThingDescriptionList{}
	res.Things = make([]*thing.ThingDescription, 0)
	resp, err := srv.store.List(args.ThingIDs, args.Limit, args.Offset)
	if err == nil {
		for _, val := range resp {
			doc := thing.ThingDescription{}
			//err = oj.Unmarshal([]byte(val), &doc, nil)
			err = json.Unmarshal([]byte(val), &doc)
			res.Things = append(res.Things, &doc)
		}
	}
	return res, err
}

func (srv *ThingKVStoreServer) Query(_ context.Context, args *svc.QueryTD_Args) (*thing.ThingDescriptionList, error) {
	res := &thing.ThingDescriptionList{}
	res.Things = make([]*thing.ThingDescription, 0)
	resp, err := srv.store.Query(args.JsonPathQuery, int(args.Limit), int(args.Offset), args.ThingIDs)
	if err == nil {
		for _, docText := range resp {
			var td thing.ThingDescription
			err = protojson.Unmarshal([]byte(docText), &td)
			res.Things = append(res.Things, &td)
		}
	}
	return res, err
}

func (srv *ThingKVStoreServer) Read(_ context.Context, args *svc.ReadTD_Args) (*thing.ThingDescription, error) {
	var res *thing.ThingDescription
	resp, err := srv.store.Read(args.ThingID)
	if err == nil {
		err = json.Unmarshal([]byte(resp), &res)
	}
	return res, err
}

func (srv *ThingKVStoreServer) Remove(_ context.Context, args *svc.RemoveTD_Args) (*emptypb.Empty, error) {
	srv.store.Remove(args.ThingID)
	return nil, nil
}

func (srv *ThingKVStoreServer) Write(_ context.Context, args *thing.ThingDescription) (*emptypb.Empty, error) {
	//tdSerialized, _ := proto.Marshal(args)
	tdSerialized, _ := json.Marshal(args)
	err := srv.store.Write(args.Id, string(tdSerialized))
	return nil, err
}

// NewThingKVStoreServer creates a service to access TDs in the state store
//  storeName is the name of the TD state store. Default is 'DefaultTDStoreName'
func NewThingKVStoreServer(thingStorePath string) (svc.ThingStoreServer, error) {

	kvStore, err := kvstore.NewKVStore(thingStorePath)
	srv := &ThingKVStoreServer{
		store: kvStore,
	}
	return srv, err
}
