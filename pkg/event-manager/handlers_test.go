///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/events"
	eventTypes "github.com/vmware/dispatch/pkg/events"
	eventsmocks "github.com/vmware/dispatch/pkg/events/mocks"
	helpers "github.com/vmware/dispatch/pkg/testing/api"
)

var testCloudEvent1 = eventTypes.CloudEvent{
	Namespace:          "dispatchframework.io",
	EventType:          "test.event",
	EventTypeVersion:   "0.1",
	CloudEventsVersion: eventTypes.CloudEventsVersion,
	SourceType:         "testsource",
	SourceID:           "testsource-id",
	EventID:            uuid.NewV4().String(),
	EventTime:          time.Now(),
	SchemaURL:          "http://some.url.com/file",
	ContentType:        "application/json",
	Extensions:         nil,
	Data:               []byte(`{"example":"value"}`),
}

func TestEventsEmitEvent(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	queue := &eventsmocks.Transport{}
	h := Handlers{es, queue, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	queue.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	reqBody := &models.Emission{
		Event: cloudEventToSwagger(&testCloudEvent1),
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
	assert.Equal(t, "test.event", *respBody.Event.EventType)
	queue.AssertCalled(t, "Publish", mock.Anything, mock.Anything, (&testCloudEvent1).DefaultTopic(), "")
}

func TestEventsEmitError(t *testing.T) {
	api := operations.NewEventManagerAPI(nil)
	es := helpers.MakeEntityStore(t)
	queue := &eventsmocks.Transport{}
	h := Handlers{es, queue, nil}
	helpers.MakeAPI(t, h.ConfigureHandlers, api)

	queue.On("Publish", mock.Anything).Return(nil)

	reqBody := &models.Emission{
		Event: &models.CloudEvent{},
	}
	r := httptest.NewRequest("POST", "/v1/event/", nil)
	params := events.EmitEventParams{
		HTTPRequest: r,
		Body:        reqBody,
	}
	responder := api.EventsEmitEventHandler.Handle(params, "testCookie")
	var respBody models.Error
	helpers.HandlerRequest(t, responder, &respBody, 400)

	assert.NotEmpty(t, respBody.Message)
	assert.Equal(t, int64(http.StatusBadRequest), respBody.Code)
}
