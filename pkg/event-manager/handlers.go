///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

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
	Config           string `long:"config" description:"Path to Config file" default:"./config.dev.json"`
	DbFile           string `long:"db-file" description:"Path to BoltDB file" default:"./db.bolt"`
	FunctionManager  string `long:"function-manager" description:"Function manager endpoint" default:"localhost:8001"`
	AMQPURL          string `long:"amqpurl" description:"URL to AMQP broker"  default:"amqp://guest:guest@localhost:5672/"`
	OrgID            string `long:"organization" description:"(temporary) Static organization id" default:"dispatch"`
	ResyncPeriod     int    `long:"resync-period" description:"The time period (in seconds) to sync with underlying k8s" default:"60"`
	K8sConfig        string `long:"kubeconfig" description:"Path to kubernetes config file" default:""`
	K8sNamespace     string `long:"namespace" description:"Kubernetes namespace" default:"default"`
	EventDriverImage string `long:"event-driver-image" description:"Event driver image"`
}{}

// Handlers is a base struct for event manager API handlers.
type Handlers struct {
	Store      entitystore.EntityStore
	EQ         events.Queue
	Controller EventController

	EventDriverController controller.Controller
	EventDriverWatcher    controller.Watcher
}

func subscriptionModelToEntity(m *models.Subscription) *Subscription {
	defer trace.Tracef("topic: %s, function: %s", *m.Topic, *m.Subscriber.Name)()
	e := Subscription{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: EventManagerFlags.OrgID,
			Name:           fmt.Sprintf("%s_%s", strings.Replace(*m.Topic, ".", "_", -1), *m.Subscriber.Name),
			Status:         entitystore.Status(m.Status),
		},
		Topic: *m.Topic,
		Subscriber: Subscriber{
			Type: *m.Subscriber.Type,
			Name: *m.Subscriber.Name,
		},
		Secrets: m.Secrets,
	}
	return &e
}

func subscriptionEntityToModel(sub *Subscription) *models.Subscription {
	defer trace.Tracef("topic: %s, function: %s", sub.Topic, sub.Subscriber)()
	m := models.Subscription{
		Name:  sub.Name,
		Topic: swag.String(sub.Topic),
		Subscriber: &models.Subscriber{
			Type: &sub.Subscriber.Type,
			Name: &sub.Subscriber.Name,
		},
		Status:       models.Status(sub.Status),
		Secrets:      sub.Secrets,
		CreatedTime:  sub.CreatedTime.Unix(),
		ModifiedTime: sub.ModifiedTime.Unix(),
	}
	return &m
}

func driverModelToEntity(m *models.Driver) *Driver {
	defer trace.Tracef("type: %s, name: %s", *m.Name, *m.Type)
	config := make(map[string]string)
	for _, c := range m.Config {
		config[c.Key] = c.Value
	}
	return &Driver{
		BaseEntity: entitystore.BaseEntity{
			OrganizationID: EventManagerFlags.OrgID,
			Name:           *m.Name,
		},
		Type:   *m.Type,
		Config: config,
	}
}

func driverEntityToModel(d *Driver) *models.Driver {
	defer trace.Tracef("type: %s, name: %s", d.Name, d.Type)

	mconfig := []*models.Config{}
	for k, v := range d.Config {
		mconfig = append(mconfig, &models.Config{Key: k, Value: v})
	}
	return &models.Driver{
		Name:         swag.String(d.Name),
		Type:         swag.String(d.Type),
		Config:       mconfig,
		Status:       models.Status(d.Status),
		CreatedTime:  d.CreatedTime.Unix(),
		ModifiedTime: d.ModifiedTime.Unix(),
	}
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

	a.ServerShutdown = func() {
		defer trace.Trace("ServerShutdown")()
		if h.EventDriverController != nil {
			h.EventDriverController.Shutdown()
		}
	}
}

func (h *Handlers) emitEvent(params eventsapi.EmitEventParams, principal interface{}) middleware.Responder {
	defer trace.Trace("emitEvent")()
	var message []byte
	var err error
	if params.Body.Payload == nil {
		message = nil
	} else {
		message, err = swag.WriteJSON(params.Body.Payload)
		if err != nil {
			return eventsapi.NewEmitEventBadRequest().WithPayload(&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("unable to parse body: %s", err)),
			})
		}
	}
	ev := events.Event{
		Topic:       *params.Body.Topic,
		ID:          uuid.NewV4().String(),
		Body:        message,
		ContentType: "application/json",
	}
	err = h.EQ.Publish(&ev)
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

func (h *Handlers) addSubscription(params subscriptionsapi.AddSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("addSubscription")()

	e := subscriptionModelToEntity(params.Body)
	e.Status = entitystore.StatusCREATING
	_, err := h.Store.Add(e)
	if err != nil {
		log.Errorf("error when storing the subscription: %+v", err)
		return eventsapi.NewEmitEventInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing the subscription"),
		})
	}
	log.Printf("updating worker...")
	h.Controller.Update(e)
	return subscriptionsapi.NewAddSubscriptionCreated().WithPayload(subscriptionEntityToModel(e))
}

func (h *Handlers) getSubscription(params subscriptionsapi.GetSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getSubscription")()
	e := Subscription{}
	err := h.Store.Get(EventManagerFlags.OrgID, params.SubscriptionName, &e)
	if err != nil {
		log.Warnf("Received GET for non-existent subscription %s", params.SubscriptionName)
		log.Debugf("store error when getting subscription: %+v", err)
		return subscriptionsapi.NewGetSubscriptionNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("subscription %s not found", params.SubscriptionName)),
			})
	}
	return subscriptionsapi.NewGetSubscriptionOK().WithPayload(subscriptionEntityToModel(&e))
}

func (h *Handlers) getSubscriptions(params subscriptionsapi.GetSubscriptionsParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getSubscriptions")()
	var subscriptions []Subscription
	err := h.Store.List(EventManagerFlags.OrgID, nil, &subscriptions)
	if err != nil {
		log.Errorf("store error when listing subscriptions: %+v", err)
		return subscriptionsapi.NewGetSubscriptionsDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting subscriptions"),
			})
	}
	var subscriptionModels []*models.Subscription
	for _, sub := range subscriptions {
		subscriptionModels = append(subscriptionModels, subscriptionEntityToModel(&sub))
	}
	return subscriptionsapi.NewGetSubscriptionsOK().WithPayload(subscriptionModels)
}

func (h *Handlers) deleteSubscription(params subscriptionsapi.DeleteSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("deleteSubscription")()
	e := Subscription{}
	err := h.Store.Get(EventManagerFlags.OrgID, params.SubscriptionName, &e)
	if err != nil {
		log.Warnf("Received GET for non-existent subscription %s", params.SubscriptionName)
		log.Debugf("store error when getting subscription: %+v", err)
		return subscriptionsapi.NewGetSubscriptionNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("subscription %s not found", params.SubscriptionName)),
			})
	}
	if e.Status == entitystore.StatusDELETING {
		log.Warnf("Attempting to delete subscription  %s which already is in DELETING state: %+v", e.Name)
		return subscriptionsapi.NewDeleteSubscriptionBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("Unable to delete subscription %s: subscription is already being deleted", e.Name)),
		})
	}
	e.Status = entitystore.StatusDELETING
	if _, err = h.Store.Update(e.Revision, &e); err != nil {
		log.Errorf("store error when deleting a subscription %s: %+v", e.Name, err)
		return subscriptionsapi.NewDeleteSubscriptionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting a subscription"),
		})
	}
	log.Debugf("Sending deleted subscription %s update to worker", e.Name)
	h.Controller.Update(&e)
	return subscriptionsapi.NewDeleteSubscriptionOK().WithPayload(subscriptionEntityToModel(&e))
}

var EventDriverTemplates = map[string]map[string]bool{
	"vcenter": map[string]bool{
		"vcenterurl": true,
	},
}

// make sure the input includes all required config values
func validateEventDriver(driver *Driver) error {
	template, ok := EventDriverTemplates[driver.Type]
	if !ok {
		return fmt.Errorf("no such driver %s", driver.Type)
	}
	for k := range template {
		if _, has := driver.Config[k]; has == false {
			return fmt.Errorf("no configuration field %s", k)
		}
	}
	return nil
}

func (h *Handlers) addDriver(params driverapi.AddDriverParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name: %s", *params.Body.Name)()

	e := driverModelToEntity(params.Body)

	// validate the driver config
	// TODO: find a better way to do the validation
	if err := validateEventDriver(e); err != nil {
		log.Errorln(err)
		return driverapi.NewAddDriverBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String("invalid event driver type or configuration"),
		})
	}

	e.Status = entitystore.StatusCREATING
	if _, err := h.Store.Add(e); err != nil {
		log.Errorf("store error when adding a new driver %s: %+v", e.Name, err)
		return driverapi.NewAddDriverInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing a new event driver"),
		})
	}
	if h.EventDriverWatcher != nil {
		h.EventDriverWatcher.OnAction(e)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return driverapi.NewAddDriverCreated().WithPayload(driverEntityToModel(e))
}

func (h *Handlers) getDriver(params driverapi.GetDriverParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDriver")()
	e := Driver{}
	err := h.Store.Get(EventManagerFlags.OrgID, params.DriverName, &e)
	if err != nil {
		log.Warnf("Received GET for non-existent driver %s", params.DriverName)
		log.Debugf("store error when getting driver: %+v", err)
		return driverapi.NewGetDriverNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("driver %s not found", params.DriverName)),
			})
	}
	return driverapi.NewGetDriverOK().WithPayload(driverEntityToModel(&e))
}

func (h *Handlers) getDrivers(params driverapi.GetDriversParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getDrivers")()
	var drivers []Driver

	// TODO: find out do we need a filter
	// filterDeleted := func(e entitystore.Entity) bool { return e.(*Subscription).Delete == false }

	// delete filter
	err := h.Store.List(EventManagerFlags.OrgID, nil, &drivers)
	if err != nil {
		log.Errorf("store error when listing drivers: %+v", err)
		return driverapi.NewGetDriverDefault(http.StatusInternalServerError).WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting drivers"),
			})
	}
	var driverModels []*models.Driver
	for _, driver := range drivers {
		driverModels = append(driverModels, driverEntityToModel(&driver))
	}
	return driverapi.NewGetDriversOK().WithPayload(driverModels)
}

func (h *Handlers) deleteDriver(params driverapi.DeleteDriverParams, principal interface{}) middleware.Responder {
	defer trace.Tracef("name '%s'", params.DriverName)()
	name := params.DriverName
	var e Driver
	if err := h.Store.Get(EventManagerFlags.OrgID, name, &e); err != nil {
		log.Errorf("store error when getting driver: %+v", err)
		return driverapi.NewDeleteDriverNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String("driver not found"),
			})
	}
	e.Status = entitystore.StatusDELETING
	if _, err := h.Store.Update(e.Revision, &e); err != nil {
		log.Errorf("store error when deleting the event driver %s: %+v", e.Name, err)
		return driverapi.NewDeleteDriverInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting an event driver"),
		})
	}
	if h.EventDriverWatcher != nil {
		h.EventDriverWatcher.OnAction(&e)
	} else {
		log.Debugf("note: the watcher is nil")
	}
	return driverapi.NewDeleteDriverOK().WithPayload(driverEntityToModel(&e))
}
