///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/client"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/drivers"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	eventsapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/events"
	"github.com/vmware/dispatch/pkg/event-manager/helpers"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions"
	"github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/events/validator"
	"github.com/vmware/dispatch/pkg/trace"
)

// Handlers is a base struct for event manager API handlers.
type Handlers struct {
	Store         entitystore.EntityStore
	Transport     events.Transport
	Watcher       controller.Watcher
	SecretsClient client.SecretsClient

	subscriptions *subscriptions.Handlers
	drivers       *drivers.Handlers
}

// ConfigureHandlers registers the function manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	a, ok := api.(*operations.EventManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		return token, nil
	}

	a.BearerAuth = func(token string) (interface{}, error) {
		// TODO: Once IAM issues signed tokens, validate them here.
		return token, nil
	}

	a.Logger = log.Printf

	h.subscriptions = subscriptions.NewHandlers(h.Store, h.Watcher)
	h.subscriptions.ConfigureHandlers(api)

	h.drivers = drivers.NewHandlers(h.Store, h.Watcher, h.SecretsClient)
	h.drivers.ConfigureHandlers(api)

	a.EventsEmitEventHandler = eventsapi.EmitEventHandlerFunc(h.emitEvent)
	a.EventsIngestEventHandler = eventsapi.IngestEventHandlerFunc(h.ingestEvent)

}

func (h *Handlers) emitEvent(params eventsapi.EmitEventParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "emitEvent")
	defer span.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		errMsg := fmt.Sprintf("Error validating event: %s", err)
		span.LogKV("validation_error", errMsg)
		return eventsapi.NewEmitEventBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(errMsg),
		})
	}

	ev := helpers.CloudEventFromAPI(&params.Body.CloudEvent)

	if err := validator.Validate(ev); err != nil {
		errMsg := fmt.Sprintf("Error validating event: %s", err)
		span.LogKV("validation_error", errMsg)
		return eventsapi.NewEmitEventBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(errMsg),
		})
	}
	err := h.Transport.Publish(ctx, ev, ev.DefaultTopic(), params.XDispatchOrg)
	if err != nil {
		errMsg := fmt.Sprintf("error when publishing a message to MQ: %+v", err)
		log.Error(errMsg)
		span.LogKV("error", errMsg)
		return eventsapi.NewEmitEventDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when emitting an event"),
		})
	}
	// TODO: Store emission in time series database
	return eventsapi.NewEmitEventOK().WithPayload(params.Body)
}

func (h *Handlers) ingestEvent(params eventsapi.IngestEventParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "ingestEvent")
	defer span.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		errMsg := fmt.Sprintf("Error validating event: %s", err)
		span.LogKV("validation_error", errMsg)
		return eventsapi.NewEmitEventBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(errMsg),
		})
	}

	// auth token is expected to match event driver UUID
	driverID := params.AuthToken

	var driverEntities []*entities.Driver

	filter := entitystore.FilterEverything()
	filter.Add(entitystore.FilterStat{
		Scope:   entitystore.FilterScopeField,
		Subject: "ID",
		Verb:    entitystore.FilterVerbEqual,
		Object:  driverID,
	})
	opts := entitystore.Options{Filter: filter}

	// TODO(karols): every event causes DB query, only to retrieve organization ID. This asks for caching.
	err := h.Store.ListGlobal(ctx, opts, &driverEntities)
	if err != nil {
		log.Errorf("error retrieving driverEntities for ID %s: %+v", driverID, err)
		return eventsapi.NewIngestEventDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when processing request"),
			})
	}

	if len(driverEntities) != 1 {
		log.Errorf("did not find driver for token %s: %+v", driverID, err)
		return eventsapi.NewIngestEventUnauthorized().WithPayload(
			&v1.Error{
				Code:    http.StatusUnauthorized,
				Message: swag.String("token not recognized"),
			})
	}

	ev := helpers.CloudEventFromAPI(&params.Body.CloudEvent)
	if err := validator.Validate(ev); err != nil {
		errMsg := fmt.Sprintf("Error validating event: %s", err)
		span.LogKV("validation_error", errMsg)
		return eventsapi.NewEmitEventBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(errMsg),
		})
	}

	// We found driver matching the auth token. We will use it to send event.
	err = h.Transport.Publish(ctx, ev, ev.DefaultTopic(), driverEntities[0].OrganizationID)
	if err != nil {
		errMsg := fmt.Sprintf("error when publishing a message to MQ: %+v", err)
		log.Error(errMsg)
		span.LogKV("error", errMsg)
		return eventsapi.NewEmitEventDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when emitting an event"),
		})
	}

	return eventsapi.NewIngestEventOK().WithPayload(params.Body)
}
