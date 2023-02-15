package core

import (
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Subscription struct {
	topic   string
	parts   []string
	handler func(topic string, message []byte)
	id      string
}

// return true if the given topic parts match the subscription
func (sub *Subscription) match(parts []string) bool {
	if len(sub.parts) != len(parts) {
		return false
	}
	for i, subPart := range sub.parts {
		if subPart != "+" {
			if subPart != parts[i] {
				return false
			}
		}
	}
	return true
}

// PubSubCore performs the actual publishing and subscription management
type PubSubCore struct {
	// list of subscribers
	subscribers []*Subscription
	submux      sync.RWMutex
}

// find the subscribers to a topic
// topic must be a full topic without wildcards
func (psc *PubSubCore) findSubscribers(topic string) (subs []*Subscription) {
	subs = make([]*Subscription, 0)
	// how many subscribers are expected? A few dozen, hundreds?
	// Right now a simple iteration is good enough.
	parts := strings.Split(topic, "/")
	for _, sub := range psc.subscribers {
		if sub.match(parts) {
			subs = append(subs, sub)
		}
	}
	return subs
}

// Publish the topic to subscribers
func (psc *PubSubCore) Publish(publisherID, topic string, message []byte) {
	subs := psc.findSubscribers(topic)
	//logrus.Infof("publisherID='%s'; topic=%v; %d subscribers", publisherID, topic, len(subs))
	for _, sub := range subs {
		sub.handler(topic, message)
	}
}

// Start a new core
func (psc *PubSubCore) Start() (err error) {
	return nil
}

// Stop ends remaining subscriptions and returns an error if subscriptions were remaining
func (psc *PubSubCore) Stop() (err error) {
	psc.submux.Lock()
	if len(psc.subscribers) > 0 {
		err = fmt.Errorf("%d subscriptions are not released. Releasing them now", len(psc.subscribers))
		logrus.Error(err)
		psc.subscribers = make([]*Subscription, 0)
	}
	psc.submux.Unlock()
	return err
}

// Subscribe to a topic
//
//	 subscriberID is the device, user or serviceID
//		topic is the topic to subscribe to. The use of '+' wildcard is supported
//		handler is the callback to invoke when a message is received
//
// This returns a subscription ID, used to unsubscribe
func (psc *PubSubCore) Subscribe(
	subscriberID string, topic string,
	handler func(topic string, message []byte)) (subscriptionID string, err error) {

	sub := &Subscription{
		topic:   topic,
		parts:   strings.Split(topic, "/"),
		handler: handler,
		id:      uuid.NewString(),
	}
	psc.submux.Lock()
	psc.subscribers = append(psc.subscribers, sub)
	psc.submux.Unlock()
	logrus.Infof("topic=%v. => subscriptionID=%s", topic, sub.id)

	return sub.id, nil
}

// Unsubscribe from one or more topics
//
//	subscriptionIDs as provided during subscribe
func (psc *PubSubCore) Unsubscribe(subscriptionIDs []string) error {
	psc.submux.Lock()
	logrus.Infof("ids=%v", subscriptionIDs)
	for _, subscriptionID := range subscriptionIDs {

		for i, sub := range psc.subscribers {
			if sub.id == subscriptionID {
				if len(psc.subscribers) > i+1 {
					// slow but keeps subscriptions in order
					psc.subscribers = append(psc.subscribers[:i], psc.subscribers[i+1:]...)
				} else {
					// remove the last one
					psc.subscribers = psc.subscribers[:len(psc.subscribers)-1]
				}
				break
			}
		}
	}
	psc.submux.Unlock()
	return nil
}

// NewPubSubCore creates a new instance of the pubsub core
func NewPubSubCore() *PubSubCore {
	psc := PubSubCore{
		subscribers: make([]*Subscription, 0),
	}
	return &psc
}
