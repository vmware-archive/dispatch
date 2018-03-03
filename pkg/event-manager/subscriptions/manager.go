///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package subscriptions

import (
	"context"
	"fmt"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/event-manager/helpers"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/entities"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/function-manager/gen/models"
	"github.com/vmware/dispatch/pkg/trace"
)

// Manager defines the subscription manager interface
type Manager interface {
	Run([]*entities.Subscription) error
	Create(context.Context, *entities.Subscription) error
	Delete(context.Context, *entities.Subscription) error
}

type defaultManager struct {
	queue    events.Transport
	fnClient client.FunctionsClient

	sync.RWMutex
	activeSubs map[string]events.Subscription
}

// NewManager creates a new subscription manager
func NewManager(mq events.Transport, fnClient client.FunctionsClient) (Manager, error) {
	defer trace.Trace("")()
	ec := defaultManager{
		queue:      mq,
		fnClient:   fnClient,
		activeSubs: make(map[string]events.Subscription),
	}

	return &ec, nil
}

func (m *defaultManager) Run(subscriptions []*entities.Subscription) error {
	defer trace.Trace("")()
	log.Debugf("event consumer initializing")

	for _, sub := range subscriptions {
		log.Debugf("Processing sub %s", sub.Name)
		m.Create(context.Background(), sub)
	}
	return nil
}

// Create creates an active subscription to Message Queue. Active subscription connects
// to Message Queue and executes a handler for every event received.
func (m *defaultManager) Create(ctx context.Context, sub *entities.Subscription) error {
	defer trace.Tracef("event %s, function %s", sub.EventType, sub.Function)()
	m.Lock()
	defer m.Unlock()
	if eventSub, ok := m.activeSubs[sub.ID]; ok {
		log.Debugf("types.Subscription for %s/%s already existed, unsubscribing", sub.EventType, sub.Function)
		eventSub.Unsubscribe()
		delete(m.activeSubs, sub.ID)
	}
	topic := fmt.Sprintf("%s.%s", sub.SourceType, sub.EventType)
	eventSub, err := m.queue.Subscribe(ctx, topic, m.handler(sub))
	if err != nil {
		err = errors.Wrapf(err, "unable to create an EventQueue subscription for event %s and function %s", sub.EventType, sub.Function)
		log.Error(err)
		return err
	}
	m.activeSubs[sub.ID] = eventSub
	return nil
}

// Delete deletes a subscription from pool of active subscriptions.
func (m *defaultManager) Delete(ctx context.Context, sub *entities.Subscription) error {
	defer trace.Tracef("event %s", sub.EventType)()
	m.Lock()
	defer m.Unlock()

	if eventSub, ok := m.activeSubs[sub.ID]; ok {
		eventSub.Unsubscribe()
		delete(m.activeSubs, sub.ID)
	}
	log.Debugf("Deleting subscription topic=%s id=%s revision=%d", sub.EventType, sub.Name, sub.Revision)
	return nil
}

// Shutdown ends event controller loop
func (m *defaultManager) Shutdown() {
	defer trace.Trace("")()
	log.Infof("Event controller shutdown")
	m.Lock()
	defer m.Unlock()
	for _, sub := range m.activeSubs {
		sub.Unsubscribe()
	}
}

// handler creates a function to handle the incoming event. it takes name of the function to be invoked as an argument.
func (m *defaultManager) handler(sub *entities.Subscription) func(context.Context, *events.CloudEvent) {
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
		m.runFunction(sub.Function, event, sub.Secrets)
	}
}

// executes a function by connecting to function manager
func (m *defaultManager) runFunction(fnName string, event *events.CloudEvent, secrets []string) {
	defer trace.Tracef("function:%s", fnName)()

	run := client.FunctionRun{}
	run.Blocking = false
	run.FunctionName = fnName
	run.Input = event.Data
	eventCopy := *event
	eventCopy.Data = ""
	run.Event = (*models.CloudEvent)(helpers.CloudEventToSwagger(&eventCopy))

	result, err := m.fnClient.RunFunction(context.Background(), &run)
	if err != nil {
		log.Warnf("Unable to run function %s, error from function manager: %+v", fnName, err)
		return
	}
	log.Debugf("Function %s returned %+v", result.FunctionName, result.Output)
	return
}
