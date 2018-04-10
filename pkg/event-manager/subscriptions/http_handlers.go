///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package subscriptions

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	subscriptionsapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/subscriptions"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/entities"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

// Handlers is a base struct for event manager API handlers.
type Handlers struct {
	orgID   string
	store   entitystore.EntityStore
	watcher controller.Watcher
}

// NewHandlers Creates new instance of subscription handlers
func NewHandlers(store entitystore.EntityStore, watcher controller.Watcher, orgID string) *Handlers {
	return &Handlers{
		watcher: watcher,
		store:   store,
		orgID:   orgID,
	}
}

// ConfigureHandlers configures API handlers for Subscription endpoints
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
	defer trace.Trace("ConfigureHandlers")()
	a, ok := api.(*operations.EventManagerAPI)
	if !ok {
		panic("Cannot configure api")
	}

	a.SubscriptionsAddSubscriptionHandler = subscriptionsapi.AddSubscriptionHandlerFunc(h.addSubscription)
	a.SubscriptionsGetSubscriptionHandler = subscriptionsapi.GetSubscriptionHandlerFunc(h.getSubscription)
	a.SubscriptionsGetSubscriptionsHandler = subscriptionsapi.GetSubscriptionsHandlerFunc(h.getSubscriptions)
	a.SubscriptionsUpdateSubscriptionHandler = subscriptionsapi.UpdateSubscriptionHandlerFunc(h.updateSubscription)
	a.SubscriptionsDeleteSubscriptionHandler = subscriptionsapi.DeleteSubscriptionHandlerFunc(h.deleteSubscription)
}

// addSubscription handles creation of new Event Subscriptions
func (h *Handlers) addSubscription(params subscriptionsapi.AddSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("addSubscription")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.addSubscription")
	defer sp.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		return subscriptionsapi.NewAddSubscriptionBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("error validating the payload: %s", err)),
		})
	}

	s := &entities.Subscription{}
	s.FromModel(params.Body, h.orgID)
	s.Status = entitystore.StatusCREATING
	_, err := h.store.Add(s)
	if err != nil {
		if entitystore.IsUniqueViolation(err) {
			return subscriptionsapi.NewAddSubscriptionConflict().WithPayload(&models.Error{
				Code:    http.StatusConflict,
				Message: swag.String("error creating subscription: non-unique name"),
			})
		}
		log.Errorf("error when storing the subscription: %+v", err)
		return subscriptionsapi.NewAddSubscriptionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when storing the subscription"),
		})
	}
	log.Printf("updating worker...")
	h.watcher.OnAction(s)
	return subscriptionsapi.NewAddSubscriptionCreated().WithPayload(s.ToModel())
}

// getSubscription handles retrieval of single Subscription
func (h *Handlers) getSubscription(params subscriptionsapi.GetSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getSubscription")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getSubscription")
	defer sp.Finish()

	s := entities.Subscription{}
	var err error

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return subscriptionsapi.NewGetSubscriptionBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	err = h.store.Get(h.orgID, params.SubscriptionName, opts, &s)
	if err != nil {
		log.Warnf("Received GET for non-existent subscription %s", params.SubscriptionName)
		log.Debugf("store error when getting subscription: %+v", err)
		return subscriptionsapi.NewGetSubscriptionNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("subscription %s not found", params.SubscriptionName)),
			})
	}
	return subscriptionsapi.NewGetSubscriptionOK().WithPayload(s.ToModel())
}

// getSubscriptions handles retrieval of Subscription list
func (h *Handlers) getSubscriptions(params subscriptionsapi.GetSubscriptionsParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getSubscriptions")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getSubscriptions")
	defer sp.Finish()

	var subscriptions []*entities.Subscription
	var err error
	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return subscriptionsapi.NewGetSubscriptionsBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}

	err = h.store.List(h.orgID, opts, &subscriptions)
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
		subscriptionModels = append(subscriptionModels, sub.ToModel())
	}
	return subscriptionsapi.NewGetSubscriptionsOK().WithPayload(subscriptionModels)
}

func (h *Handlers) updateSubscription(params subscriptionsapi.UpdateSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("updateSubscription")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.updateSubscription")
	defer sp.Finish()

	s := &entities.Subscription{}
	var err error

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return subscriptionsapi.NewUpdateSubscriptionBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	err = h.store.Get(h.orgID, params.SubscriptionName, opts, s)
	if err != nil {
		log.Warnf("Received UPDATE for non-existent subscription %s", params.SubscriptionName)
		log.Debugf("store error when getting subscription: %+v", err)
		return subscriptionsapi.NewUpdateSubscriptionNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(err.Error()),
			})
	}
	if s.Status == entitystore.StatusUPDATING {
		log.Warnf("Attempting to update subscription %s which already is in UPDATING state: %+v", s.Name)
		return subscriptionsapi.NewUpdateSubscriptionBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Unable to update subscription %s: subscription is already being updated", s.Name)),
			})
	}

	s.FromModel(params.Body, h.orgID)
	s.Status = entitystore.StatusUPDATING
	if _, err = h.store.Update(s.Revision, s); err != nil {
		log.Errorf("store error when updating a subscription %s: %+v", s.Name, err)
		return subscriptionsapi.NewUpdateSubscriptionInternalServerError().WithPayload(
			&models.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when updating a subscription"),
			})
	}
	log.Debugf("Sending updated subscription %s update to worker", s.Name)
	h.watcher.OnAction(s)
	return subscriptionsapi.NewUpdateSubscriptionOK().WithPayload(s.ToModel())
}

// deleteSubscription handles deletion of a Subscription
func (h *Handlers) deleteSubscription(params subscriptionsapi.DeleteSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("deleteSubscription")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.deleteSubscription")
	defer sp.Finish()

	s := &entities.Subscription{}
	var err error

	opts := entitystore.Options{
		Filter: entitystore.FilterEverything(),
	}
	opts.Filter, err = utils.ParseTags(opts.Filter, params.Tags)
	if err != nil {
		log.Errorf(err.Error())
		return subscriptionsapi.NewDeleteSubscriptionBadRequest().WithPayload(
			&models.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(err.Error()),
			})
	}
	err = h.store.Get(h.orgID, params.SubscriptionName, opts, s)
	if err != nil {
		log.Warnf("Received DELETE for non-existent subscription %s", params.SubscriptionName)
		log.Debugf("store error when getting subscription: %+v", err)
		return subscriptionsapi.NewDeleteSubscriptionNotFound().WithPayload(
			&models.Error{
				Code:    http.StatusNotFound,
				Message: swag.String(fmt.Sprintf("subscription %s not found", params.SubscriptionName)),
			})
	}
	if s.Status == entitystore.StatusDELETING {
		log.Warnf("Attempting to delete subscription  %s which already is in DELETING state: %+v", s.Name)
		return subscriptionsapi.NewDeleteSubscriptionBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("Unable to delete subscription %s: subscription is already being deleted", s.Name)),
		})
	}
	s.Status = entitystore.StatusDELETING
	if _, err = h.store.Update(s.Revision, s); err != nil {
		log.Errorf("store error when deleting a subscription %s: %+v", s.Name, err)
		return subscriptionsapi.NewDeleteSubscriptionInternalServerError().WithPayload(&models.Error{
			Code:    http.StatusInternalServerError,
			Message: swag.String("internal server error when deleting a subscription"),
		})
	}
	log.Debugf("Sending deleted subscription %s update to worker", s.Name)
	h.watcher.OnAction(s)
	return subscriptionsapi.NewDeleteSubscriptionOK().WithPayload(s.ToModel())
}
