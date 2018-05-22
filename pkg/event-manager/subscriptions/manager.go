///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package subscriptions

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/dispatch/pkg/api/v1"

	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/event-manager/helpers"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/entities"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

// Manager defines the subscription manager interface
type Manager interface {
	Run(context.Context, []*entities.Subscription) error
	Create(context.Context, *entities.Subscription) error
	Update(context.Context, *entities.Subscription) error
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
	ec := defaultManager{
		queue:      mq,
		fnClient:   fnClient,
		activeSubs: make(map[string]events.Subscription),
	}

	return &ec, nil
}

func (m *defaultManager) Run(ctx context.Context, subscriptions []*entities.Subscription) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()
	log.Debugf("event consumer initializing")

	for _, sub := range subscriptions {
		log.Debugf("Processing sub %s", sub.Name)
		m.Create(ctx, sub)
	}
	return nil
}

// Create creates an active subscription to Message Queue. Active subscription connects
// to Message Queue and executes a handler for every event received.
func (m *defaultManager) Create(ctx context.Context, sub *entities.Subscription) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	span.SetTag("eventType", sub.EventType)
	span.SetTag("functionName", sub.Function)

	m.Lock()
	defer m.Unlock()
	if eventSub, ok := m.activeSubs[sub.ID]; ok {
		log.Debugf("types.Subscription for %s/%s already existed, unsubscribing", sub.EventType, sub.Function)
		eventSub.Unsubscribe()
		delete(m.activeSubs, sub.ID)
	}
	topic := fmt.Sprintf("%s.%s", sub.SourceType, sub.EventType)
	eventSub, err := m.queue.Subscribe(ctx, topic, m.handler(ctx, sub))
	if err != nil {
		err = errors.Wrapf(err, "unable to create a subscription for event %s and function %s", sub.EventType, sub.Function)
		span.LogKV("error", err)
		log.Error(err)
		return err
	}
	m.activeSubs[sub.ID] = eventSub
	return nil
}

// Update updates a subscription
func (m *defaultManager) Update(ctx context.Context, sub *entities.Subscription) error {
	return m.Create(ctx, sub)
}

// Delete deletes a subscription from pool of active subscriptions.
func (m *defaultManager) Delete(ctx context.Context, sub *entities.Subscription) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	span.SetTag("eventType", sub.EventType)
	span.SetTag("functionName", sub.Function)

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
	log.Infof("Event controller shutdown")
	m.Lock()
	defer m.Unlock()
	for _, sub := range m.activeSubs {
		sub.Unsubscribe()
	}
}

// handler creates a function to handle the incoming event. it takes name of the function to be invoked as an argument.
func (m *defaultManager) handler(ctx context.Context, sub *entities.Subscription) func(context.Context, *events.CloudEvent) {
	span, _ := trace.Trace(ctx, "")
	defer span.Finish()

	span.SetTag("eventType", sub.EventType)
	span.SetTag("functionName", sub.Function)

	return func(ctx context.Context, event *events.CloudEvent) {
		span, ctx := trace.Trace(ctx, "EventHandler")
		defer span.Finish()
		span.SetTag("eventType", sub.EventType)
		span.SetTag("functionName", sub.Function)

		m.runFunction(ctx, sub.Function, event, sub.Secrets)
	}
}

// executes a function by connecting to function manager
func (m *defaultManager) runFunction(ctx context.Context, fnName string, event *events.CloudEvent, secrets []string) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	span.SetTag("eventType", event.EventType)
	span.SetTag("functionName", fnName)

	processedData, err := m.processEventData(event)
	if err != nil {
		log.Warnf("Unable to process event payload: %+v", err)
	}

	run := v1.Run{
		Blocking:     false,
		FunctionName: fnName,
		Input:        processedData,
	}
	eventCopy := *event
	eventCopy.Data = ""
	run.Event = helpers.CloudEventToAPI(&eventCopy)

	result, err := m.fnClient.RunFunction(ctx, &run)
	if err != nil {
		errorMsg := fmt.Sprintf("Unable to run function %s, error from function manager: %+v", fnName, err)
		span.LogKV("error", errorMsg)
		log.Error(errorMsg)
		return
	}
	span.LogKV("functionName", result.FunctionName,
		"functionResult", result.Output)
	log.Debugf("Function %s returned %+v", result.FunctionName, result.Output)

	return
}

func (m *defaultManager) processEventData(event *events.CloudEvent) (interface{}, error) {
	switch event.ContentType {
	case "application/json":
		var jsonData interface{}
		err := json.Unmarshal([]byte(event.Data), &jsonData)
		if err != nil {
			return nil, errors.Wrap(err, "error when unmarshaling JSON data")
		}
		return jsonData, nil
	default:
		// for every other content type we pass data as is
		return event.Data, nil
	}
}
