///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"context"
	"fmt"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

// SubscriptionManager defines the subscription manager interface
type SubscriptionManager interface {
	Run([]*Subscription) error
	Create(context.Context, *Subscription) error
	Delete(context.Context, *Subscription) error
}

type subscriptionManager struct {
	queue    events.Transport
	fnClient client.FunctionsClient

	sync.RWMutex
	activeSubs map[string]events.Subscription
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager(mq events.Transport, fnClient client.FunctionsClient) (SubscriptionManager, error) {
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
		ec.Create(context.Background(), sub)
	}
	return nil
}

// Create creates an active subscription to Message Queue. Active subscription connects
// to Message Queue and executes a handler for every event received.
func (ec *subscriptionManager) Create(ctx context.Context, sub *Subscription) error {
	defer trace.Tracef("event %s, function %s", sub.EventType, sub.Function)()
	ec.Lock()
	defer ec.Unlock()
	if eventSub, ok := ec.activeSubs[sub.ID]; ok {
		log.Debugf("Subscription for %s/%s already existed, unsubscribing", sub.EventType, sub.Function)
		eventSub.Unsubscribe()
		delete(ec.activeSubs, sub.ID)
	}
	topic := fmt.Sprintf("%s.%s.%s", sub.SourceType, sub.SourceName, sub.EventType)
	eventSub, err := ec.queue.Subscribe(ctx, topic, ec.handler(sub))
	if err != nil {
		err = errors.Wrapf(err, "unable to create an EventQueue subscription for event %s and function %s", sub.EventType, sub.Function)
		log.Error(err)
		return err
	}
	ec.activeSubs[sub.ID] = eventSub
	return nil
}

// Delete deletes a subscription from pool of active subscriptions.
func (ec *subscriptionManager) Delete(ctx context.Context, sub *Subscription) error {
	defer trace.Tracef("event %s", sub.EventType)()
	ec.Lock()
	defer ec.Unlock()

	if eventSub, ok := ec.activeSubs[sub.ID]; ok {
		eventSub.Unsubscribe()
		delete(ec.activeSubs, sub.ID)
	}
	log.Debugf("Deleting subscription topic=%s id=%s revision=%d", sub.EventType, sub.Name, sub.Revision)
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
func (ec *subscriptionManager) handler(sub *Subscription) func(context.Context, *events.CloudEvent) {
	defer trace.Tracef("function name:%s", sub.Function)()

	return func(ctx context.Context, event *events.CloudEvent) {
		trace.Tracef("HandlerClosure(). function name:%s, event:%s", sub.Name, event.EventID)()
		sp, _ := opentracing.StartSpanFromContext(
			ctx,
			"EventManager.EventHandler",
			opentracing.Tag{Key: "subscriptionName", Value: sub.Name},
			opentracing.Tag{Key: "eventID", Value: event.EventID},
		)
		defer sp.Finish()

		// TODO: Pass tracing context once Function Manager is tracing-aware
		ec.runFunction(sub.Function, event, sub.Secrets)
	}
}

// executes a function by connecting to function manager
func (ec *subscriptionManager) runFunction(fnName string, event *events.CloudEvent, secrets []string) {
	defer trace.Tracef("function:%s", fnName)()

	run := client.FunctionRun{}
	run.Blocking = true
	run.FunctionName = fnName
	run.Input = event

	result, err := ec.fnClient.RunFunction(context.Background(), &run)
	if err != nil {
		log.Warnf("Unable to run function %s, error from function manager: %+v", fnName, err)
		return
	}
	log.Debugf("Function %s returned %+v", result.FunctionName, result.Output)
	return
}
