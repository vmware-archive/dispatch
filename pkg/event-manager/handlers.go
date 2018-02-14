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
	"github.com/vmware/dispatch/pkg/events/validator"
	"github.com/vmware/dispatch/pkg/utils"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	driverapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/drivers"
	eventsapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/events"
	subscriptionsapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/subscriptions"
	events "github.com/vmware/dispatch/pkg/events"
	"github.com/vmware/dispatch/pkg/trace"
)

// EventManagerFlags are configuration flags for the function manager
var EventManagerFlags = struct {
	Config            string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile            string `long:"db-file" description:"Backend DB URL/Path" default:"./db.bolt"`
	DbBackend         string `long:"db-backend" description:"Backend DB Name" default:"boltdb"`
	DbUser            string `long:"db-username" description:"Backend DB Username" default:"dispatch"`
	DbPassword        string `long:"db-password" description:"Backend DB Password" default:"dispatch"`
	DbDatabase        string `long:"db-database" description:"Backend DB Name" default:"dispatch"`
	FunctionManager   string `long:"function-manager" description:"Function manager endpoint" default:"localhost:8001"`
	Transport         string `long:"transport" description:"Event transport to use" default:"rabbitmq"`
	RabbitMQURL       string `long:"rabbitmq-url" description:"URL to RabbitMQ broker" default:"amqp://guest:guest@localhost:5672/"`
	OrgID             string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
	ResyncPeriod      int    `long:"resync-period" description:"The time period (in seconds) to sync with underlying k8s" default:"60"`
	K8sConfig         string `long:"kubeconfig" description:"Path to kubernetes config file" default:""`
	K8sNamespace      string `long:"namespace" description:"Kubernetes namespace" default:"default"`
	EventDriverImage  string `long:"event-driver-image" description:"Default event driver image"`
	EventSidecarImage string `long:"event-sidecar-image" description:"Event sidecar image"`
	SecretStore       string `long:"secret-store" description:"Secret store endpoint" default:"localhost:8003"`
	TracerURL         string `long:"tracer-url" description:"Open Tracing Tracer URL" default:""`
}{}

// Handlers is a base struct for event manager API handlers.
type Handlers struct {
	Store   entitystore.EntityStore
	EQ      events.Transport
	Watcher controller.Watcher
}

// ConfigureHandlers registers the function manager handlers to the API
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	defer trace.Trace("ConfigureHandlers")()
	a, ok := api.(*operations.EventManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.CookieAuth = func(token string) (interface{}, error) {
		// TODO: be able to retrieve user information from the cookie
		// currently just return the cookie
		log.Printf("cookie auth: %s\n", token)
		return token, nil
	}

	a.Logger = log.Printf
	a.EventsEmitEventHandler = eventsapi.EmitEventHandlerFunc(h.emitEvent)
	a.SubscriptionsAddSubscriptionHandler = subscriptionsapi.AddSubscriptionHandlerFunc(h.addSubscription)
	a.SubscriptionsGetSubscriptionHandler = subscriptionsapi.GetSubscriptionHandlerFunc(h.getSubscription)
	a.SubscriptionsGetSubscriptionsHandler = subscriptionsapi.GetSubscriptionsHandlerFunc(h.getSubscriptions)
	a.SubscriptionsDeleteSubscriptionHandler = subscriptionsapi.DeleteSubscriptionHandlerFunc(h.deleteSubscription)
	a.DriversAddDriverHandler = driverapi.AddDriverHandlerFunc(h.addDriver)
	a.DriversGetDriverHandler = driverapi.GetDriverHandlerFunc(h.getDriver)
	a.DriversGetDriversHandler = driverapi.GetDriversHandlerFunc(h.getDrivers)
	a.DriversDeleteDriverHandler = driverapi.DeleteDriverHandlerFunc(h.deleteDriver)
	a.DriversAddDriverTypeHandler = driverapi.AddDriverTypeHandlerFunc(h.addDriverType)
	a.DriversGetDriverTypeHandler = driverapi.GetDriverTypeHandlerFunc(h.getDriverType)
	a.DriversGetDriverTypesHandler = driverapi.GetDriverTypesHandlerFunc(h.getDriverTypes)
	a.DriversDeleteDriverTypeHandler = driverapi.DeleteDriverTypeHandlerFunc(h.deleteDriverType)

}

func (h *Handlers) emitEvent(params eventsapi.EmitEventParams, principal interface{}) middleware.Responder {
	defer trace.Trace("emitEvent")()

	sp, spCtx := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.emitEvent")
	defer sp.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		return eventsapi.NewEmitEventBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("Error validating event: %s", err)),
		})
	}

	ev := cloudEventFromSwagger(params.Body.Event)

	if err := validator.Validate(ev); err != nil {
		return eventsapi.NewEmitEventBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("Error validating event: %s", err)),
		})
	}

	err := h.EQ.Publish(spCtx, ev, ev.DefaultTopic(), "")
	if err != nil {
		log.Errorf("error when publishing a message to MQ: %+v", err)
		return eventsapi.NewEmitEventInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when emitting an event"),
		})
	}
	// TODO: Store emission in time series database
	return eventsapi.NewEmitEventOK().WithPayload(params.Body)
}
