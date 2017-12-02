///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/client"
	clientmocks "github.com/vmware/dispatch/pkg/client/mocks"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/events"
	eventsmocks "github.com/vmware/dispatch/pkg/events/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func mockEventController(es entitystore.EntityStore, queue events.Queue, fnClient client.FunctionsClient) *eventController {
	return &eventController{
		queue:        queue,
		store:        es,
		activeSubs:   make(map[string]events.Subscription),
		subChan:      make(chan *Subscription),
		fnClient:     fnClient,
		resyncPeriod: DefaultResyncPeriod,
	}
}

func TestControllerRun(t *testing.T) {
	fnClient := &clientmocks.FunctionsClient{}
	queue := &eventsmocks.Queue{}
	es := helpers.MakeEntityStore(t)

	controller, err := NewEventController(es, queue, fnClient, ResyncPeriod(10*time.Second))
	assert.NoError(t, err)

	assert.NoError(t, controller.Run())

	controller.Shutdown()
}

func TestControllerRunWithSubs(t *testing.T) {
	fnClient := &clientmocks.FunctionsClient{}
	queue := &eventsmocks.Queue{}
	es := helpers.MakeEntityStore(t)

	controller, err := NewEventController(es, queue, fnClient)
	assert.NoError(t, err)

	subscription := &eventsmocks.Subscription{}

	sub := &Subscription{
		BaseEntity: entitystore.BaseEntity{
			Name:   "sub1",
			Status: entitystore.StatusCREATING,
		},
		Topic: "test.topic",
		Subscriber: Subscriber{
			Type: FunctionSubscriber,
			Name: "test.function",
		},
	}

	es.Add(sub)
	subscription.On("Unsubscribe").Return(nil)
	queue.On("Subscribe", sub.Topic, mock.AnythingOfType("events.Handler")).Return(subscription, nil).Once()
	assert.NoError(t, controller.Run())
	controller.Shutdown()
	queue.AssertNumberOfCalls(t, "Subscribe", 1)
}

func TestControllerWatchEvents(t *testing.T) {
	fnClient := &clientmocks.FunctionsClient{}
	queue := &eventsmocks.Queue{}
	es := helpers.MakeEntityStore(t)

	// Used to test watch() which runs as go routine
	wg := sync.WaitGroup{}

	controller := mockEventController(es, queue, fnClient)
	assert.NoError(t, controller.Run())

	sub := &Subscription{
		BaseEntity: entitystore.BaseEntity{
			Name:           "sub1",
			Status:         entitystore.StatusCREATING,
			OrganizationID: EventManagerFlags.OrgID,
		},
		Topic: "test.topic",
		Subscriber: Subscriber{
			Type: FunctionSubscriber,
			Name: "test.function",
		},
	}
	es.Add(sub)
	subscription := &eventsmocks.Subscription{}
	subscription.On("Unsubscribe").Return(nil).Run(
		func(_ mock.Arguments) {
			wg.Done() // Unsubscribe is called after lock is acquired
		},
	)

	queue.On("Subscribe", sub.Topic, mock.AnythingOfType("events.Handler")).Return(subscription, nil).Once().Run(
		func(_ mock.Arguments) {
			wg.Done() // Subscribe is called by createSub, after lock is acquired
		},
	)

	// The workflow here is as follows:
	// 1. Update(sub) sends a message to a channel and returns
	// 2. wg.Wait() blocks until wg.Done() is called
	// 3. wg.Done() is called by Subscribe() mock. it's called inside the createSub() function, which already holds lock
	// 4. controller.Lock() will block until createSub() finishes (and releases the lock)
	// 5. activeSubs will have the subscription added by createSub() at the time controller.Lock() acquires the lock.
	wg.Add(1)
	controller.Update(sub)
	wg.Wait()
	controller.Lock()
	assert.Len(t, controller.activeSubs, 1)
	controller.Unlock()

	err := es.Get(EventManagerFlags.OrgID, sub.Name, sub)
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)
	assert.Equal(t, entitystore.StatusREADY, sub.Status)

	// same workflow, just uses Unsubscribe() to call wg.Done()
	sub.Status = entitystore.StatusDELETING
	_, err = es.Update(sub.Revision, sub)

	if err != nil {
		panic(err)
	}

	wg.Add(1)
	controller.Update(sub)
	wg.Wait()
	controller.Lock()
	assert.Len(t, controller.activeSubs, 0)
	controller.Unlock()

	controller.Shutdown()

	var subs []*Subscription
	err = es.List(EventManagerFlags.OrgID, nil, &subs)
	assert.NoError(t, err)
	assert.Len(t, subs, 0)

	queue.AssertNumberOfCalls(t, "Subscribe", 1)
	subscription.AssertNumberOfCalls(t, "Unsubscribe", 1)
}

func TestControllerSync(t *testing.T) {
	EventManagerFlags.OrgID = "dispatch"
	fnClient := &clientmocks.FunctionsClient{}
	queue := &eventsmocks.Queue{}
	es := helpers.MakeEntityStore(t)

	// Used to test watch() which runs as go routine
	wg := sync.WaitGroup{}

	controller := mockEventController(es, queue, fnClient)
	controller.resyncPeriod = 1 * time.Second
	assert.NoError(t, controller.Run())

	sub := &Subscription{
		BaseEntity: entitystore.BaseEntity{
			Name:           "sub1",
			Status:         entitystore.StatusCREATING,
			OrganizationID: EventManagerFlags.OrgID,
		},
		Topic: "test.topic",
		Subscriber: Subscriber{
			Type: FunctionSubscriber,
			Name: "test.function",
		},
	}
	es.Add(sub)
	subscription := &eventsmocks.Subscription{}
	subscription.On("Unsubscribe").Return(nil).Run(
		func(_ mock.Arguments) {
			wg.Done() // Unsubscribe is called after lock is acquired
		},
	)

	queue.On("Subscribe", sub.Topic, mock.AnythingOfType("events.Handler")).Return(subscription, nil).Once().Run(
		func(_ mock.Arguments) {
			wg.Done() // Subscribe is called by createSub, after lock is acquired
		},
	)

	wg.Add(1)
	wg.Wait()
	controller.Lock()
	assert.Len(t, controller.activeSubs, 1)
	controller.Unlock()

	err := es.Get(EventManagerFlags.OrgID, sub.Name, sub)
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)
	assert.Equal(t, entitystore.StatusREADY, sub.Status)

	// same workflow, just uses Unsubscribe() to call wg.Done()
	sub.Status = entitystore.StatusDELETING
	_, err = es.Update(sub.Revision, sub)

	if err != nil {
		panic(err)
	}

	wg.Add(1)
	wg.Wait()
	controller.Lock()
	assert.Len(t, controller.activeSubs, 0)
	controller.Unlock()

	controller.Shutdown()

	var subs []*Subscription
	err = es.List(EventManagerFlags.OrgID, nil, &subs)
	assert.NoError(t, err)
	assert.Len(t, subs, 0)

	queue.AssertNumberOfCalls(t, "Subscribe", 1)
	subscription.AssertNumberOfCalls(t, "Unsubscribe", 1)
}

func TestRunFunction(t *testing.T) {
	fnClient := &clientmocks.FunctionsClient{}
	queue := &eventsmocks.Queue{}
	es := helpers.MakeEntityStore(t)
	controller := mockEventController(es, queue, fnClient)
	ev := &events.Event{
		Topic:       "test.topic",
		Body:        []byte("{\"key\": \"value\"}"),
		ContentType: "application/json",
		ID:          uuid.NewV4().String(),
	}
	fnClient.On("RunFunction", mock.Anything, mock.AnythingOfType("*client.FunctionRun")).Return(&client.FunctionRun{}, nil).Once()
	controller.runFunction("testFunction", ev, []string{"secret1", "secret2"})

	fnClient.On("RunFunction", mock.Anything, mock.AnythingOfType("*client.FunctionRun")).Return(&client.FunctionRun{}, errors.New("testerror")).Once()
	controller.runFunction("testFunction", ev, nil)
	fnClient.AssertNumberOfCalls(t, "RunFunction", 2)
}

func TestEmitEvent(t *testing.T) {
	fnClient := &clientmocks.FunctionsClient{}
	queue := &eventsmocks.Queue{}
	es := helpers.MakeEntityStore(t)
	controller := mockEventController(es, queue, fnClient)
	ev := &events.Event{
		Topic:       "test.topic",
		Body:        []byte("{\"key\": \"value\"}"),
		ContentType: "application/json",
		ID:          uuid.NewV4().String(),
	}
	queue.On("Publish", mock.AnythingOfType("*events.Event")).Return(nil).Once()
	controller.emitEvent("test.topic.new", ev)

	queue.On("Publish", mock.AnythingOfType("*events.Event")).Return(errors.New("testerror")).Once()
	controller.emitEvent("test.topic.new", ev)
	queue.AssertNumberOfCalls(t, "Publish", 2)
}
