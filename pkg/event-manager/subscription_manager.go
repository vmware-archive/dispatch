///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

type SubscriptionManager interface {
	Run([]*Subscription) error
	Create(*Subscription) error
	Delete(*Subscription) error
}

type subscriptionManager struct {
	queue    events.Queue
	fnClient client.FunctionsClient

	sync.RWMutex
	activeSubs map[string]events.Subscription
}

func NewSubscriptionManager(mq events.Queue, fnClient client.FunctionsClient) (SubscriptionManager, error) {
	defer trace.Trace("")()
	ec := subscriptionManager{
		queue:      mq,
		fnClient:   fnClient,
		activeSubs: make(map[string]events.Subscription),
	}

	return &ec, nil
}

func (ec *subscriptionManager) Run(subscriptions []*Subscription) error {
	defer trace.Trace("")()
	log.Debugf("event consumer initializing")

	for _, sub := range subscriptions {
		log.Debugf("Processing sub %s", sub.Name)
		ec.Create(sub)
	}
	return nil
}

func (ec *subscriptionManager) Create(sub *Subscription) error {
	defer trace.Tracef("event %s, function %s", sub.Topic, sub.Subscriber)()
	ec.Lock()
	defer ec.Unlock()
	if eventSub, ok := ec.activeSubs[sub.ID]; ok {
		log.Debugf("Subscription for %s/%s already existed, unsubscribing", sub.Topic, sub.Subscriber)
		eventSub.Unsubscribe()
		delete(ec.activeSubs, sub.ID)
	}

	eventSub, err := ec.queue.Subscribe(sub.Topic, ec.handler(sub))
	if err != nil {
		err = errors.Wrapf(err, "unable to create an EventQueue subscription for event %s and function %s", sub.Topic, sub.Subscriber)
		log.Error(err)
		return err
	}
	ec.activeSubs[sub.ID] = eventSub
	return nil
}

func (ec *subscriptionManager) Delete(sub *Subscription) error {
	defer trace.Tracef("event %s", sub.Topic)()
	ec.Lock()
	defer ec.Unlock()

	if eventSub, ok := ec.activeSubs[sub.ID]; ok {
		eventSub.Unsubscribe()
		delete(ec.activeSubs, sub.ID)
	}
	log.Debugf("Deleting subscription topic=%s id=%s revision=%d", sub.Topic, sub.Name, sub.Revision)
	return nil
}

// Shutdown ends event controller loop
func (ec *subscriptionManager) Shutdown() {
	defer trace.Trace("")()
	log.Infof("Event controller shutdown")
	ec.Lock()
	defer ec.Unlock()
	for _, sub := range ec.activeSubs {
		sub.Unsubscribe()
	}
}

// handler creates a function to handle the incoming event. it takes name of the function to be invoked as an argument.
func (ec *subscriptionManager) handler(sub *Subscription) func(*events.Event) {
	defer trace.Tracef("subscriber type:%s, name:%s", sub.Subscriber.Type, sub.Subscriber.Name)()
	return func(event *events.Event) {
		trace.Tracef("HandlerClosure(). subscriber type:%s, name:%s, event:", sub.Subscriber.Type, sub.Subscriber.Name, event.ID)()

		switch sub.Subscriber.Type {
		case FunctionSubscriber:
			ec.runFunction(sub.Subscriber.Name, event, sub.Secrets)
		case EventSubscriber:
			ec.emitEvent(sub.Subscriber.Name, event)
		}
	}
}

func (ec *subscriptionManager) runFunction(fnName string, event *events.Event, secrets []string) {
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

	result, err := ec.fnClient.RunFunction(context.Background(), &run)
	if err != nil {
		log.Warnf("Unable to run function %s, error from function manager: %+v", fnName, err)
		return
	}
	log.Debugf("Function %s returned %+v", result.FunctionName, result.Output)
	return
}

func (ec *subscriptionManager) emitEvent(newTopic string, event *events.Event) {
	defer trace.Tracef("oldTopic:%s,newTopic:%s", event.Topic, newTopic)()
	log.Debug("converting event from %s to %s", event.Topic, newTopic)
	newEvent := events.Event{
		Topic:       newTopic,
		ID:          uuid.NewV4().String(),
		Body:        event.Body,
		ContentType: event.ContentType,
	}
	if err := ec.queue.Publish(&newEvent); err != nil {
		log.Warnf("Error when publishing on topic %s: %+v", newTopic, err)
	}
}
