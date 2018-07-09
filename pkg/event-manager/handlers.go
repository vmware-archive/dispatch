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
	"github.com/vmware/dispatch/pkg/client"

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/drivers"
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
