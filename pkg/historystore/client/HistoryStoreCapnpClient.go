// Package client that wraps the capnp generated client with a POGS API
package client

import (
	"context"
	"net"
	"time"

	"capnproto.org/go/capnp/v3/rpc"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/caphelp"
)

// HistoryStoreCapnpClient provides a POGS wrapper around the capnp client API
// This implements the IHistoryStore interface
type HistoryStoreCapnpClient struct {
	connection *rpc.Conn           // connection to capnp server
	capability hubapi.HistoryStore // capnp client
	ctx        context.Context
	ctxCancel  context.CancelFunc
}

// AddAction adds a Thing action with the given name and value to the action history
func (cl *HistoryStoreCapnpClient) AddAction(ctx context.Context, actionValue thing.ThingValue) error {

	method, release := cl.capability.AddAction(cl.ctx,
		func(params hubapi.HistoryStore_addAction_Params) error {
			capValue := caphelp.ThingValueToCapnp(actionValue)
			err2 := params.SetActionValue(capValue)
			return err2
		})
	defer release()
	_, err := method.Struct()
	return err
}

// AddEvent adds an event to the event history
func (cl *HistoryStoreCapnpClient) AddEvent(ctx context.Context, eventValue thing.ThingValue) error {

	method, release := cl.capability.AddEvent(cl.ctx,
		func(params hubapi.HistoryStore_addEvent_Params) error {
			capValue := caphelp.ThingValueToCapnp(eventValue)
			err2 := params.SetEventValue(capValue)
			return err2
		})
	defer release()
	_, err := method.Struct()
	return err
}

func (cl *HistoryStoreCapnpClient) AddEvents(ctx context.Context, events []thing.ThingValue) error {

	method, release := cl.capability.AddEvents(cl.ctx,
		func(params hubapi.HistoryStore_addEvents_Params) error {
			capValues := caphelp.ThingValueListToCapnp(events)
			err2 := params.SetEventValues(capValues)
			return err2
		})
	defer release()
	_, err := method.Struct()
	return err
}

// GetActionHistory returns the history of a Thing action
// before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)
func (cl *HistoryStoreCapnpClient) GetActionHistory(ctx context.Context,
	thingID string, actionName string, after string, before string, limit int) (
	values []thing.ThingValue, err error) {

	method, release := cl.capability.GetActionHistory(cl.ctx,
		func(params hubapi.HistoryStore_getActionHistory_Params) error {
			err2 := params.SetThingID(thingID)
			_ = params.SetActionName(actionName)
			_ = params.SetAfter(after)
			_ = params.SetBefore(before)
			params.SetLimit(int32(limit))
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capValues, _ := resp.Values()
		values = caphelp.CapnpToThingValueList(capValues)
	}
	return values, err
}

// GetEventHistory returns the history of a Thing event
// before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)
func (cl *HistoryStoreCapnpClient) GetEventHistory(ctx context.Context,
	thingID string, eventName string, after string, before string, limit int) (
	values []thing.ThingValue, err error) {

	method, release := cl.capability.GetEventHistory(cl.ctx,
		func(params hubapi.HistoryStore_getEventHistory_Params) error {
			err2 := params.SetThingID(thingID)
			_ = params.SetEventName(eventName)
			_ = params.SetAfter(after)
			_ = params.SetBefore(before)
			params.SetLimit(int32(limit))
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capValues, _ := resp.Values()
		values = caphelp.CapnpToThingValueList(capValues)
	}
	return values, err
}

// GetLatestEvents returns a map of the latest event values of a Thing
func (cl *HistoryStoreCapnpClient) GetLatestEvents(ctx context.Context, thingID string) (
	latest map[string]thing.ThingValue, err error) {

	method, release := cl.capability.GetLatestEvents(cl.ctx,
		func(params hubapi.HistoryStore_getLatestEvents_Params) error {
			err2 := params.SetThingID(thingID)
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capValueMap, _ := resp.ThingValueMap()
		latest = caphelp.CapnpToThingValueMap(capValueMap)
	}
	return latest, err
}

func (cl *HistoryStoreCapnpClient) Info(ctx context.Context) (info StoreInfo, err error) {

	method, release := cl.capability.Info(cl.ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capInfo, _ := resp.Statistics()
		engine, _ := capInfo.Engine()
		nrActions := capInfo.NrActions()
		nrEvents := capInfo.NrEvents()
		uptimeSec := capInfo.Uptime()
		info = StoreInfo{
			Engine:    engine,
			NrActions: int(nrActions),
			NrEvents:  int(nrEvents),
			Uptime:    int(uptimeSec),
		}
	}
	return info, err
}

// NewHistoryStoreCapnpClient returns a history store client using the capnp protocol
func NewHistoryStoreCapnpClient(address string, isUDS bool) (*HistoryStoreCapnpClient, error) {
	var cl *HistoryStoreCapnpClient
	network := "tcp"
	if isUDS {
		network = "unix"
	}
	connection, err := net.Dial(network, address)
	if err == nil {
		transport := rpc.NewStreamTransport(connection)
		rpcConn := rpc.NewConn(transport, nil)
		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*60)
		capability := hubapi.HistoryStore(rpcConn.Bootstrap(ctx))

		cl = &HistoryStoreCapnpClient{
			connection: rpcConn,
			capability: capability,
			ctx:        ctx,
			ctxCancel:  ctxCancel,
		}
	}
	return cl, nil
}
