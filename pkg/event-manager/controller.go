///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/client"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

const DefaultResyncPeriod = 60 * time.Second

// EventController watches for new events and processes them
type EventController interface {
	Run() error
	Update(*Subscription)
	Shutdown()
}

// ResyncPeriod sets the duration between state resynchronization calls
func ResyncPeriod(period time.Duration) func(ec *eventController) error {
	return func(ec *eventController) error {
		// TODO: implement resync
		ec.resyncPeriod = period
		return nil
	}
}

// eventController is an implementation of EventController interface.
type eventController struct {
	resyncPeriod time.Duration
	queue        events.Queue
	store        entitystore.EntityStore
	subChan      chan *Subscription
	done         chan struct{}
	fnClient     client.FunctionsClient

	sync.RWMutex
	activeSubs map[string]events.Subscription
}

// NewEventController creates new instance of EventController, which handles the execution of subscriptions
func NewEventController(es entitystore.EntityStore, mq events.Queue, fnClient client.FunctionsClient, options ...func(ec *eventController) error) (EventController, error) {
	defer trace.Trace("")()
	ec := eventController{
		queue:        mq,
		store:        es,
		activeSubs:   make(map[string]events.Subscription),
		subChan:      make(chan *Subscription),
		fnClient:     fnClient,
		resyncPeriod: DefaultResyncPeriod,
	}

	for _, option := range options {
		option(&ec)
	}

	return &ec, nil

}

// Run creates missing subs and starts EventController loop
func (ew *eventController) Run() error {
	defer trace.Trace("")()
	log.Debugf("event controller initializing")
	if ew.done != nil {
		return errors.New("event controller already initialized")
	}

	subscriptions, err := ew.getSubs(false)
	if err != nil {
		err = errors.Wrap(err, "error initializing event controller")
		log.Error(err)
		return err
	}

	ew.done = make(chan struct{})

	var created, deleted int
	for _, sub := range subscriptions {
		log.Debugf("Processing sub %s", sub.Name)
		if sub.Status == entitystore.StatusREADY || sub.Status == entitystore.StatusCREATING {
			ew.createSub(&sub)
			created++
		} else if sub.Status == entitystore.StatusDELETING {
			ew.deleteSub(&sub)
			deleted++
		}
	}
	log.Infof("Created %d subscriptions", created)
	log.Infof("Deleted %d subscriptions", deleted)
	go ew.watch()
	return nil
}

// Update adds Subscription change to the queue
func (ew *eventController) Update(sub *Subscription) {
	defer trace.Tracef("event: %s", sub.Topic)()
	ew.subChan <- sub
}

// Shutdown ends event controller loop
func (ew *eventController) Shutdown() {
	defer trace.Trace("")()
	log.Infof("Event controller shutdown")
	close(ew.done)
	ew.Lock()
	defer ew.Unlock()
	for _, sub := range ew.activeSubs {
		sub.Unsubscribe()
	}
	ew.done = nil
}

// watch runs an infinite loop waiting for either new subscription to be created/deleted, or a shutting down signal to be sent.
func (ew *eventController) watch() {
	defer trace.Trace("")()
	for {
		select {
		case sub := <-ew.subChan:
			if sub.Status == entitystore.StatusDELETING {
				log.Debugf("received subscription %s/%s to delete", sub.OrganizationID, sub.Name)
				ew.deleteSub(sub)
			} else if sub.Status == entitystore.StatusCREATING {
				log.Debugf("received subscription %s/%s to add", sub.OrganizationID, sub.Name)
				ew.createSub(sub)
			} else {
				log.Warnf("unexpected subscription %s in state %s received", sub.Name, sub.Status)
			}
		case <-time.After(ew.resyncPeriod):
			log.Debugf("periodic sync triggered")
			ew.sync()
		case <-ew.done:
			log.Debugf("Shutting down eventController watch loop")
			return
		}
	}

}

// getSubs returns the list of current subscriptions in database. set inProgress to true to only return subscriptions in DELETING or CREATING state.
func (ew *eventController) getSubs(inProgress bool) ([]Subscription, error) {
	defer trace.Trace("")()

	var subscriptions []Subscription
	filter := func(entity entitystore.Entity) bool {
		sub := entity.(*Subscription)
		return !inProgress || sub.Status == entitystore.StatusDELETING || sub.Status == entitystore.StatusCREATING
	}
	err := ew.store.List(EventManagerFlags.OrgID, filter, &subscriptions)
	if err != nil {
		err = errors.Wrap(err, "store error when listing subscriptions")
		log.Error(err)
		return nil, err
	}
	return subscriptions, nil
}

func (ew *eventController) createSub(sub *Subscription) error {
	defer trace.Tracef("event %s, function %s", sub.Topic, sub.Subscriber)()
	ew.Lock()
	defer ew.Unlock()
	if eventSub, ok := ew.activeSubs[sub.ID]; ok {
		log.Debugf("Subscription for %s/%s already existed, unsubscribing", sub.Topic, sub.Subscriber)
		eventSub.Unsubscribe()
		delete(ew.activeSubs, sub.ID)
	}

	eventSub, err := ew.queue.Subscribe(sub.Topic, ew.handler(sub))
	if err != nil {
		err = errors.Wrapf(err, "unable to create an EventQueue subscription for event %s and function %s", sub.Topic, sub.Subscriber)
		log.Error(err)
		return err
	}
	ew.activeSubs[sub.ID] = eventSub

	if sub.Status == entitystore.StatusREADY {
		// no need to update status in DB
		return nil
	}

	sub.Status = entitystore.StatusREADY
	_, err = ew.store.Update(sub.Revision, sub)
	// TODO: handle case where revision changed between store.List and store.Update (run store.Get again)
	if err != nil {
		log.Errorf("error when updating subscription %s status: %+v", sub.Name, err)
		log.Debugf("removing subscription %s from active subs", sub.Name)
		eventSub.Unsubscribe()
		return err
	}
	return nil
}

func (ew *eventController) deleteSub(sub *Subscription) error {
	defer trace.Tracef("deleteSub for event %s", sub.Topic)()
	ew.Lock()
	defer ew.Unlock()

	if eventSub, ok := ew.activeSubs[sub.ID]; ok {
		eventSub.Unsubscribe()
		delete(ew.activeSubs, sub.ID)
	}
	log.Debugf("Deleting subscription topic=%s id=%s revision=%d", sub.Topic, sub.Name, sub.Revision)
	err := ew.store.Delete(EventManagerFlags.OrgID, sub.Name, sub)
	if err != nil {
		log.Errorf("error when deleting subscription %s: %+v", sub.Name, err)
		return err
	}
	return nil
}

func (ew *eventController) sync() error {
	subscriptions, err := ew.getSubs(true)
	if err != nil {
		err = errors.Wrap(err, "error retrieving subscriptions for sync purposes")
		log.Error(err)
		return err
	}

	var created, deleted int
	for _, sub := range subscriptions {
		log.Debugf("Processing sub %s", sub.Name)
		if sub.Status == entitystore.StatusCREATING {
			err = ew.createSub(&sub)
			if err != nil {
				return errors.Wrap(err, "error when creating a sub in synchronization process")
			}
			created++
		} else if sub.Status == entitystore.StatusDELETING {
			ew.deleteSub(&sub)
			if err != nil {
				return errors.Wrap(err, "error when deleting a sub in synchronization process")
			}
			deleted++
		}
	}
	log.Debugf("Sync task created %d and deleted %d subscriptions", created, deleted)
	return nil
}

// handler creates a function to handle the incoming event. it takes name of the function to be invoked as an argument.
func (ew *eventController) handler(sub *Subscription) func(*events.Event) {
	defer trace.Tracef("subscriber type:%s, name:%s", sub.Subscriber.Type, sub.Subscriber.Name)()
	return func(event *events.Event) {
		trace.Tracef("HandlerClosure(). subscriber type:%s, name:%s, event:", sub.Subscriber.Type, sub.Subscriber.Name, event.ID)()

		switch sub.Subscriber.Type {
		case FunctionSubscriber:
			ew.runFunction(sub.Subscriber.Name, event, sub.Secrets)
		case EventSubscriber:
			ew.emitEvent(sub.Subscriber.Name, event)
		}
	}
}

func (ew *eventController) runFunction(fnName string, event *events.Event, secrets []string) {
	defer trace.Tracef("function:%s", fnName)()
	var input map[string]interface{}

	err := json.Unmarshal(event.Body, &input)
	if err != nil {
		log.Warnf("Unable to run  %s, error parsing payload: %+v", fnName, err)
		return
	}
	run := client.FunctionRun{}
	run.Blocking = true
	run.FunctionName = fnName
	run.Input = input

	result, err := ew.fnClient.RunFunction(context.Background(), &run)
	if err != nil {
		log.Warnf("Unable to run function %s, error from function manager: %+v", fnName, err)
		return
	}
	log.Debugf("Function %s returned %+v", result.FunctionName, result.Output)
	return
}

func (ew *eventController) emitEvent(newTopic string, event *events.Event) {
	defer trace.Tracef("oldTopic:%s,newTopic:%s", event.Topic, newTopic)()
	log.Debug("converting event from %s to %s", event.Topic, newTopic)
	newEvent := events.Event{
		Topic:       newTopic,
		ID:          uuid.NewV4().String(),
		Body:        event.Body,
		ContentType: event.ContentType,
	}
	if err := ew.queue.Publish(&newEvent); err != nil {
		log.Warnf("Error when publishing on topic %s: %+v", newTopic, err)
	}
}
