// Package capnpclient that wraps the capnp generated client with a POGS API
package capnpclient

import (
	"context"

	"github.com/hiveot/hub.capnp/go/hubapi"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub/internal/caphelp"
	"github.com/hiveot/hub/pkg/history"
)

// ReadHistoryCapnpClient provides a POGS wrapper around the capnp client API
// This implements the IReadHistory interface
type ReadHistoryCapnpClient struct {
	capability hubapi.CapReadHistory // capnp client
}

func (cl *ReadHistoryCapnpClient) Release() {
	cl.capability.Release()
}

// GetActionHistory returns the history of a Thing action
// before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)
func (cl *ReadHistoryCapnpClient) GetActionHistory(ctx context.Context,
	thingID string, actionName string, after string, before string, limit int) (
	values []thing.ThingValue, err error) {

	method, release := cl.capability.GetActionHistory(ctx,
		func(params hubapi.CapReadHistory_getActionHistory_Params) error {
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
		values = caphelp.UnmarshalThingValueList(capValues)
	}
	return values, err
}

// GetEventHistory returns the history of a Thing event
// before and after are timestamps in iso8601 format (YYYY-MM-DDTHH:MM:SS-TZ)
func (cl *ReadHistoryCapnpClient) GetEventHistory(ctx context.Context,
	thingID string, eventName string, after string, before string, limit int) (
	values []thing.ThingValue, err error) {

	method, release := cl.capability.GetEventHistory(ctx,
		func(params hubapi.CapReadHistory_getEventHistory_Params) error {
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
		values = caphelp.UnmarshalThingValueList(capValues)
	}
	return values, err
}

// GetLatestEvents returns a map of the latest event values of a Thing
func (cl *ReadHistoryCapnpClient) GetLatestEvents(
	ctx context.Context, thingID string) (
	latest map[string]thing.ThingValue, err error) {

	method, release := cl.capability.GetLatestEvents(ctx,
		func(params hubapi.CapReadHistory_getLatestEvents_Params) error {
			err2 := params.SetThingID(thingID)
			return err2
		})
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capValueMap, _ := resp.ThingValueMap()
		latest = caphelp.UnmarshalThingValueMap(capValueMap)
	}
	return latest, err
}

func (cl *ReadHistoryCapnpClient) Info(
	ctx context.Context) (info history.StoreInfo, err error) {

	method, release := cl.capability.Info(ctx, nil)
	defer release()
	resp, err := method.Struct()
	if err == nil {
		capInfo, _ := resp.Statistics()
		engine, _ := capInfo.Engine()
		nrActions := capInfo.NrActions()
		nrEvents := capInfo.NrEvents()
		uptimeSec := capInfo.Uptime()
		info = history.StoreInfo{
			Engine:    engine,
			NrActions: int(nrActions),
			NrEvents:  int(nrEvents),
			Uptime:    int(uptimeSec),
		}
	}
	return info, err
}

// NewReadHistoryCapnpClient returns a read history client using the capnp protocol
// Intended for internal use.
func NewReadHistoryCapnpClient(cap hubapi.CapReadHistory) *ReadHistoryCapnpClient {
	cl := &ReadHistoryCapnpClient{capability: cap}
	return cl
}
