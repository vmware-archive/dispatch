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

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/gen/models"
	eventsapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/events"
	subscriptionsapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/subscriptions"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

func (h *Handlers) addSubscription(params subscriptionsapi.AddSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("addSubscription")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.addSubscription")
	defer sp.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		return eventsapi.NewEmitEventBadRequest().WithPayload(&models.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("error validating the payload: %s", err)),
		})
	}

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
	h.Watcher.OnAction(e)
	return subscriptionsapi.NewAddSubscriptionCreated().WithPayload(subscriptionEntityToModel(e))
}

func (h *Handlers) getSubscription(params subscriptionsapi.GetSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("getSubscription")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getSubscription")
	defer sp.Finish()

	e := Subscription{}
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
	err = h.Store.Get(EventManagerFlags.OrgID, params.SubscriptionName, opts, &e)
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

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.getSubscriptions")
	defer sp.Finish()

	var subscriptions []*Subscription
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

	err = h.Store.List(EventManagerFlags.OrgID, opts, &subscriptions)
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
		subscriptionModels = append(subscriptionModels, subscriptionEntityToModel(sub))
	}
	return subscriptionsapi.NewGetSubscriptionsOK().WithPayload(subscriptionModels)
}

func (h *Handlers) deleteSubscription(params subscriptionsapi.DeleteSubscriptionParams, principal interface{}) middleware.Responder {
	defer trace.Trace("deleteSubscription")()

	sp, _ := utils.AddHTTPTracing(params.HTTPRequest, "EventManager.deleteSubscription")
	defer sp.Finish()

	e := Subscription{}
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
	err = h.Store.Get(EventManagerFlags.OrgID, params.SubscriptionName, opts, &e)
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
	h.Watcher.OnAction(&e)
	return subscriptionsapi.NewDeleteSubscriptionOK().WithPayload(subscriptionEntityToModel(&e))
}
