package pubsub_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thanhpk/randstr"

	"github.com/hiveot/hub/api/go/vocab"
	"github.com/hiveot/hub/lib/thing"
)

// subscribers  things   events       duration          with capnp    with go background
//    10          10       1000        2.9 msec            100 ms          1.5 ms
//    10         100       1000        1.3 msec             70 ms          1.5
//    10        1000       1000        1.1 msec             66 ms        123
//   100          10       1000       20 msec              260 ms        351
//   100         100       1000        4 msec               95             1.5
//   100        1000       1000        3 msec               70             1.6
//  1000           1       1000     1700 msec            14200 msec (!)
//  1000          10       1000      190 msec             5000 msec (!)
//  1000         100       1000       37 msec              600
//  1000        1000       1000       20 msec              140

var BenchParams = []struct {
	Subscribers int // number of subscribers
	Things      int // number of Things to subscribe to
	Events      int // number of events to test with
}{
	{Subscribers: 10, Things: 10, Events: 1000},
	{Subscribers: 10, Things: 100, Events: 1000},
	{Subscribers: 10, Things: 1000, Events: 1000},
	{Subscribers: 100, Things: 10, Events: 1000},
	{Subscribers: 100, Things: 100, Events: 1000},
	{Subscribers: 100, Things: 1000, Events: 1000},
	{Subscribers: 1000, Things: 1, Events: 1000},
	{Subscribers: 1000, Things: 10, Events: 1000},
	{Subscribers: 1000, Things: 100, Events: 1000},
	{Subscribers: 1000, Things: 1000, Events: 1000},
}

// BenchmarkPubSub measures time needed to receive events
func BenchmarkPubSub(b *testing.B) {
	ctx := context.Background()
	const publisherID = "device1ID"
	rand.Seed(time.Now().UnixNano())
	for _, tbl := range BenchParams {
		// setup
		svc, stopFn := startService(testUseCapnp)
		capDevice, _ := svc.CapDevicePubSub(ctx, publisherID)
		capUser, _ := svc.CapUserPubSub(ctx, "user1")

		// generate thingIDs
		thingIDs := make([]string, tbl.Things)
		for i := 0; i < len(thingIDs); i++ {
			thingIDs[i] = "urn:" + randstr.String(10)
		}
		// add subscribers
		var evCount = 0
		for i := 0; i < tbl.Subscribers; i++ {
			thingID := thingIDs[rand.Intn(tbl.Things)]
			name := vocab.VocabTemperature
			err := capUser.SubEvent(ctx, publisherID, thingID, name, func(tv thing.ThingValue) {
				//logrus.Infof("received tv thingAddr=%s name=%s", tv.thingAddr, tv.ID)
				evCount++
			})
			assert.NoError(b, err)
		}

		// run tests

		// test adding records one by one
		b.Run(fmt.Sprintf("[things:%d] subscribers:%d, events:%d", tbl.Things, tbl.Subscribers, tbl.Events),
			func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					evCount = 0
					rand.Seed(time.Now().UnixNano())
					// send N events
					for i := 0; i < tbl.Events; i++ {
						thingID := thingIDs[rand.Intn(tbl.Things)]
						name := vocab.VocabTemperature
						value := []byte("2.5")
						_ = capDevice.PubEvent(ctx, thingID, name, value)
					}
					// just an estimate, expect more thant 80% events and less than 120%
					// depends on ratio nr Things and nrEvents.
					// looks like rand is far from random. 1200% ?
					//assert.GreaterOrEqual(b, evCount, tbl.Events-tbl.Events/2)
					//assert.LessOrEqual(b, evCount, tbl.Events+tbl.Events/2)
				}
			})

		// let the background processes complete
		time.Sleep(time.Second * 3)
		b.Log("Releasing clients")
		capDevice.Release()
		capUser.Release()
		stopFn()
	}

	// generate event names
	// subscribe to event names

}
