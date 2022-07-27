package thingstore

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/wostzone/wost.grpc/go/svc"
	"github.com/wostzone/wost.grpc/go/thing"

	dapr "github.com/dapr/go-sdk/client"
)

const DefaultTDStoreName = "wost.things"

// ThingStoreServer implements the svc.ThingStoreServer interface
// TD's are kept in memory for quick response and search indexing
type ThingStoreServer struct {
	svc.UnimplementedTDStoreServer
	stateClient dapr.Client
	storeName   string
	// name of the field containing the index list of Things. Used for ListTD
	thingIndexName string
}

func (srv *ThingStoreServer) ListTD(ctx context.Context, args *svc.ListTD_Args) (*thing.ThingDescriptionList, error) {
	res := &thing.ThingDescriptionList{}
	res.Things = make([]*thing.ThingDescription, 0)

	query := `{}`
	resp, err := srv.stateClient.QueryStateAlpha1(ctx, srv.storeName, query, nil)
	if err != nil {
		return nil, err
	}
	for _, queryItem := range resp.Results {
		td := &thing.ThingDescription{}
		err = proto.Unmarshal(queryItem.Value, td)
		res.Things = append(res.Things, td)
	}
	return res, err
}

func (srv *ThingStoreServer) ReadTD(ctx context.Context, args *svc.ReadTD_Args) (*thing.ThingDescription, error) {
	res := &thing.ThingDescription{}
	stateItem, err := srv.stateClient.GetState(ctx, srv.storeName, args.ThingID, nil)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(stateItem.Value, res)
	return res, err
}

func (srv *ThingStoreServer) RemoveTD(ctx context.Context, args *svc.RemoveTD_Args) (*emptypb.Empty, error) {
	err := srv.stateClient.DeleteState(ctx, srv.storeName, args.ThingID, nil)
	return nil, err
}

func (srv *ThingStoreServer) WriteTD(ctx context.Context, args *thing.ThingDescription) (*emptypb.Empty, error) {
	thingID := args.Id
	tdSerialized, _ := proto.Marshal(args)
	err := srv.stateClient.SaveState(ctx, srv.storeName, thingID, tdSerialized, nil)

	// Add the TD to the index
	//srv.stateClient.GetState(ctx, srv.storeName, srv.ThingIndexName, nil)

	return nil, err
}

// NewThingStoreServer creates a service to access TD's in the state store
//  storeName is the name of the TD state store. Default is 'DefaultTDStoreName'
func NewThingStoreServer(storeName string) (*ThingStoreServer, error) {
	if storeName == "" {
		storeName = DefaultTDStoreName
	}
	stateClient, err := dapr.NewClientWithSocket("/tmp/dapr-wost-test-grpc.socket")
	srv := &ThingStoreServer{
		storeName:   storeName,
		stateClient: stateClient,
	}
	return srv, err
}
