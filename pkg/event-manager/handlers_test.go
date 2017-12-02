///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/events"

	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/subscriptions"
	eventsmocks "github.com/vmware/dispatch/pkg/events/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func addSubscriptionEntity(t *testing.T, api *operations.EventManagerAPI, topic, subscriberType, subscriberName string) *models.Subscription {
	reqBody := &models.Subscription{
		Topic: swag.String(topic),
		Subscriber: &models.Subscriber{
			Type: swag.String(subscriberType),
			Name: swag.String(subscriberName),
		},
	}
	r := httptest.NewRequest("POST", "/v1/event/subscriptions", nil)
	params := subscriptions.AddSubscriptionParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.SubscriptionsAddSubscriptionHandler.Handle(params, "testCookie")
	var respBody models.Subscription
	helpers.HandlerRequest(t, responder, &respBody, 201)
	return &respBody
}

func TestEventsEmitEvent(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	queue := &eventsmocks.Queue{}
	h := Handlers{es, queue, nil, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	queue.On("Publish", mock.Anything).Return(nil)

	reqBody := &models.Emission{
		Topic: swag.String("test.topic"),
		Payload: map[string]string{
			"key": "value",
		},
	}
	r := httptest.NewRequest("POST", "/v1/event/", nil)
	params := events.EmitEventParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.EventsEmitEventHandler.Handle(params, "testCookie")
	var respBody models.Emission
	helpers.HandlerRequest(t, responder, &respBody, 200)

	assert.NotEmpty(t, respBody.ID)
	assert.Equal(t, "test.topic", *respBody.Topic)
	queue.AssertCalled(t, "Publish", mock.Anything)
}

func TestSubscriptionsAddSubscriptionHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	controller := &EventControllerMock{}
	h := Handlers{es, nil, controller, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	controller.On("Update", mock.Anything).Return()

	respBody := addSubscriptionEntity(t, api, "test.topic", FunctionSubscriber, "testfunction")

	assert.NotNil(t, respBody.CreatedTime)
	assert.NotEmpty(t, respBody.ID)
	assert.Equal(t, "test.topic", *respBody.Topic)
	assert.Equal(t, FunctionSubscriber, *respBody.Subscriber.Type)
	assert.Equal(t, "testfunction", *respBody.Subscriber.Name)
	controller.AssertCalled(t, "Update", mock.Anything)
}

func TestSubscriptionsGetSubscriptionHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	controller := &EventControllerMock{}
	h := Handlers{es, nil, controller, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	controller.On("Update", mock.Anything).Return()

	addBody := addSubscriptionEntity(t, api, "test.topic", FunctionSubscriber, "testfunction")
	assert.NotEmpty(t, addBody.ID)

	createdTime := addBody.CreatedTime
	r := httptest.NewRequest("GET", "/v1/event/subscriptions/test_topic_testfunction", nil)
	get := subscriptions.GetSubscriptionParams{
		HTTPRequest:      r,
		SubscriptionName: "test_topic_testfunction",
	}
	getResponder := api.SubscriptionsGetSubscriptionHandler.Handle(get, "testCookie")
	var getBody models.Subscription
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, addBody.ID, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, "test.topic", *getBody.Topic)
	assert.Equal(t, FunctionSubscriber, *getBody.Subscriber.Type)
	assert.Equal(t, "testfunction", *getBody.Subscriber.Name)
	controller.AssertCalled(t, "Update", mock.Anything)

	r = httptest.NewRequest("GET", "/v1/event/subscriptions/doesNotExist", nil)
	get = subscriptions.GetSubscriptionParams{
		HTTPRequest:      r,
		SubscriptionName: "doesNotExist",
	}
	getResponder = api.SubscriptionsGetSubscriptionHandler.Handle(get, "testCookie")

	var errorBody models.Error
	helpers.HandlerRequest(t, getResponder, &errorBody, 404)
	assert.EqualValues(t, http.StatusNotFound, errorBody.Code)
}

func TestSubscriptionsDeleteSubscriptionHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	controller := &EventControllerMock{}
	h := Handlers{es, nil, controller, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	controller.On("Update", mock.Anything).Return()

	addBody := addSubscriptionEntity(t, api, "test.topic", FunctionSubscriber, "testfunction")
	assert.NotEmpty(t, addBody.ID)

	r := httptest.NewRequest("GET", "/v1/event/subscriptions", nil)
	get := subscriptions.GetSubscriptionsParams{
		HTTPRequest: r,
	}
	getResponder := api.SubscriptionsGetSubscriptionsHandler.Handle(get, "testCookie")
	var getBody []models.Subscription
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 1)

	r = httptest.NewRequest("DELETE", "/v1/image/base/test_topic_testfunction", nil)
	del := subscriptions.DeleteSubscriptionParams{
		HTTPRequest:      r,
		SubscriptionName: "test_topic_testfunction",
	}
	delResponder := api.SubscriptionsDeleteSubscriptionHandler.Handle(del, "testCookie")
	var delBody models.Subscription
	helpers.HandlerRequest(t, delResponder, &delBody, 200)
	assert.Equal(t, "test_topic_testfunction", delBody.Name)

	getResponder = api.SubscriptionsGetSubscriptionsHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
}
