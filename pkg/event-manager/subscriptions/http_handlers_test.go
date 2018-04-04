///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package subscriptions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"

	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/subscriptions"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

func addSubscriptionEntity(t *testing.T, api *operations.EventManagerAPI, name, eventType, function string) *models.Subscription {
	reqBody := &models.Subscription{
		Name:       swag.String(name),
		EventType:  swag.String(eventType),
		Function:   swag.String(function),
		SourceType: swag.String("dispatch"),
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

func addSubscriptionEntityWithError(t *testing.T, api *operations.EventManagerAPI, eventType, function string) *models.Error {
	reqBody := &models.Subscription{
		EventType: swag.String(eventType),
		Function:  swag.String(function),
	}
	r := httptest.NewRequest("POST", "/v1/event/subscriptions", nil)
	params := subscriptions.AddSubscriptionParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.SubscriptionsAddSubscriptionHandler.Handle(params, "testCookie")
	var respBody models.Error
	helpers.HandlerRequest(t, responder, &respBody, 400)
	return &respBody
}

func TestSubscriptionsAddSubscriptionHandlerError(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{"", es, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addSubscriptionEntityWithError(t, api, "test.topic", "testfunction")
	assert.NotEmpty(t, respBody.Message)
	assert.Equal(t, int64(http.StatusBadRequest), respBody.Code)
}

func TestSubscriptionsAddSubscriptionHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{"", es, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	respBody := addSubscriptionEntity(t, api, "mysubscription", "test.topic", "testfunction")
	assert.Equal(t, "test.topic", *respBody.EventType)
	assert.Equal(t, "testfunction", *respBody.Function)
}

func TestSubscriptionsGetSubscriptionHandler(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	h := Handlers{"", es, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addSubscriptionEntity(t, api, "mysubscription", "test.topic", "testfunction")
	assert.NotEmpty(t, addBody.ID)

	createdTime := addBody.CreatedTime
	r := httptest.NewRequest("GET", "/v1/event/subscriptions/mysubscription", nil)
	get := subscriptions.GetSubscriptionParams{
		HTTPRequest:      r,
		SubscriptionName: "mysubscription",
	}
	getResponder := api.SubscriptionsGetSubscriptionHandler.Handle(get, "testCookie")
	var getBody models.Subscription
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Equal(t, addBody.ID, getBody.ID)
	assert.Equal(t, createdTime, getBody.CreatedTime)
	assert.Equal(t, "test.topic", *getBody.EventType)
	assert.Equal(t, "testfunction", *getBody.Function)

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
	h := Handlers{"", es, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	addBody := addSubscriptionEntity(t, api, "mysubscription", "test.topic", "testfunction")
	assert.NotEmpty(t, addBody.ID)

	r := httptest.NewRequest("GET", "/v1/event/subscriptions", nil)
	get := subscriptions.GetSubscriptionsParams{
		HTTPRequest: r,
	}
	getResponder := api.SubscriptionsGetSubscriptionsHandler.Handle(get, "testCookie")
	var getBody []models.Subscription
	helpers.HandlerRequest(t, getResponder, &getBody, 200)

	assert.Len(t, getBody, 1)

	r = httptest.NewRequest("DELETE", "/v1/event/subscriptions/mysubscription", nil)
	del := subscriptions.DeleteSubscriptionParams{
		HTTPRequest:      r,
		SubscriptionName: "mysubscription",
	}
	delResponder := api.SubscriptionsDeleteSubscriptionHandler.Handle(del, "testCookie")
	var delBody models.Subscription
	helpers.HandlerRequest(t, delResponder, &delBody, 200)
	assert.Equal(t, "mysubscription", *delBody.Name)

	getResponder = api.SubscriptionsGetSubscriptionsHandler.Handle(get, "testCookie")
	helpers.HandlerRequest(t, getResponder, &getBody, 200)
}
