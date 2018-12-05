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

	"github.com/vmware/dispatch/pkg/api/v1"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations"
	subscriptionsapi "github.com/vmware/dispatch/pkg/event-manager/gen/restapi/operations/subscriptions"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/entities"
	"github.com/vmware/dispatch/pkg/trace"
	"github.com/vmware/dispatch/pkg/utils"
)

const (
	testOrgID = "testOrg"
)

// Handlers is a base struct for event manager API handlers.
type Handlers struct {
	store   entitystore.EntityStore
	watcher controller.Watcher
}

// NewHandlers Creates new instance of subscription handlers
func NewHandlers(store entitystore.EntityStore, watcher controller.Watcher) *Handlers {
	return &Handlers{
		watcher: watcher,
		store:   store,
	}
}

// ConfigureHandlers configures API handlers for Subscription endpoints
func (h *Handlers) ConfigureHandlers(api middleware.RoutableAPI) {
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
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "addSubscription")
	defer span.Finish()

	if err := params.Body.Validate(strfmt.Default); err != nil {
		return subscriptionsapi.NewAddSubscriptionBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("error validating the payload: %s", err)),
		})
	}

	s := &entities.Subscription{}
	s.FromModel(params.Body, params.XDispatchOrg)
	s.Status = entitystore.StatusCREATING
	_, err := h.store.Add(ctx, s)
	if err != nil {
		if entitystore.IsUniqueViolation(err) {
			return subscriptionsapi.NewAddSubscriptionConflict().WithPayload(&v1.Error{
				Code:    http.StatusConflict,
				Message: utils.ErrorMsgAlreadyExists("subscription", s.Name),
			})
		}
		log.Errorf("error when storing the subscription: %+v", err)
		return subscriptionsapi.NewAddSubscriptionDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("subscription", s.Name),
		})
	}
	log.Printf("updating worker...")
	h.watcher.OnAction(ctx, s)
	return subscriptionsapi.NewAddSubscriptionCreated().WithPayload(s.ToModel())
}

// getSubscription handles retrieval of single Subscription
func (h *Handlers) getSubscription(params subscriptionsapi.GetSubscriptionParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "getSubscription")
	defer span.Finish()

	s := entities.Subscription{}
	var err error

	err = h.store.Get(ctx, params.XDispatchOrg, params.SubscriptionName, entitystore.Options{}, &s)
	if err != nil {
		log.Warnf("Received GET for non-existent subscription %s", params.SubscriptionName)
		log.Debugf("store error when getting subscription: %+v", err)
		return subscriptionsapi.NewGetSubscriptionNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("subscription", params.SubscriptionName),
			})
	}
	return subscriptionsapi.NewGetSubscriptionOK().WithPayload(s.ToModel())
}

// getSubscriptions handles retrieval of Subscription list
func (h *Handlers) getSubscriptions(params subscriptionsapi.GetSubscriptionsParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "getSubscriptions")
	defer span.Finish()

	var subscriptions []*entities.Subscription
	var err error

	err = h.store.List(ctx, params.XDispatchOrg, entitystore.Options{}, &subscriptions)
	if err != nil {
		log.Errorf("store error when listing subscriptions: %+v", err)
		return subscriptionsapi.NewGetSubscriptionsDefault(http.StatusInternalServerError).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: swag.String("internal server error when getting subscriptions"),
			})
	}
	var subscriptionModels []*v1.Subscription
	for _, sub := range subscriptions {
		subscriptionModels = append(subscriptionModels, sub.ToModel())
	}
	return subscriptionsapi.NewGetSubscriptionsOK().WithPayload(subscriptionModels)
}

func (h *Handlers) updateSubscription(params subscriptionsapi.UpdateSubscriptionParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "updateSubscription")
	defer span.Finish()

	s := &entities.Subscription{}
	var err error

	err = h.store.Get(ctx, params.XDispatchOrg, params.SubscriptionName, entitystore.Options{}, s)
	if err != nil {
		log.Warnf("Received UPDATE for non-existent subscription %s", params.SubscriptionName)
		log.Debugf("store error when getting subscription: %+v", err)
		return subscriptionsapi.NewUpdateSubscriptionNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("subscription", params.SubscriptionName),
			})
	}
	if s.Status == entitystore.StatusUPDATING {
		log.Warnf("Attempting to update subscription %s which already is in UPDATING state", s.Name)
		return subscriptionsapi.NewUpdateSubscriptionBadRequest().WithPayload(
			&v1.Error{
				Code:    http.StatusBadRequest,
				Message: swag.String(fmt.Sprintf("Unable to update subscription %s: subscription is already being updated", s.Name)),
			})
	}

	s.FromModel(params.Body, s.OrganizationID)
	s.Status = entitystore.StatusUPDATING
	if _, err = h.store.Update(ctx, s.Revision, s); err != nil {
		log.Errorf("store error when updating a subscription %s: %+v", s.Name, err)
		return subscriptionsapi.NewUpdateSubscriptionDefault(500).WithPayload(
			&v1.Error{
				Code:    http.StatusInternalServerError,
				Message: utils.ErrorMsgInternalError("subscription", s.Name),
			})
	}
	log.Debugf("Sending updated subscription %s update to worker", s.Name)
	h.watcher.OnAction(ctx, s)
	return subscriptionsapi.NewUpdateSubscriptionOK().WithPayload(s.ToModel())
}

// deleteSubscription handles deletion of a Subscription
func (h *Handlers) deleteSubscription(params subscriptionsapi.DeleteSubscriptionParams, principal interface{}) middleware.Responder {
	span, ctx := trace.Trace(params.HTTPRequest.Context(), "deleteSubscription")
	defer span.Finish()

	s := &entities.Subscription{}
	var err error

	err = h.store.Get(ctx, params.XDispatchOrg, params.SubscriptionName, entitystore.Options{}, s)
	if err != nil {
		log.Warnf("Received DELETE for non-existent subscription %s", params.SubscriptionName)
		log.Debugf("store error when getting subscription: %+v", err)
		return subscriptionsapi.NewDeleteSubscriptionNotFound().WithPayload(
			&v1.Error{
				Code:    http.StatusNotFound,
				Message: utils.ErrorMsgNotFound("subscription", params.SubscriptionName),
			})
	}
	if s.Status == entitystore.StatusDELETING {
		log.Warnf("Attempting to delete subscription  %s which already is in DELETING state", s.Name)
		return subscriptionsapi.NewDeleteSubscriptionBadRequest().WithPayload(&v1.Error{
			Code:    http.StatusBadRequest,
			Message: swag.String(fmt.Sprintf("Unable to delete subscription %s: subscription is already being deleted", s.Name)),
		})
	}
	s.Status = entitystore.StatusDELETING
	if _, err = h.store.Update(ctx, s.Revision, s); err != nil {
		log.Errorf("store error when deleting a subscription %s: %+v", s.Name, err)
		return subscriptionsapi.NewDeleteSubscriptionDefault(500).WithPayload(&v1.Error{
			Code:    http.StatusInternalServerError,
			Message: utils.ErrorMsgInternalError("subscription", s.Name),
		})
	}
	log.Debugf("Sending deleted subscription %s update to worker", s.Name)
	h.watcher.OnAction(ctx, s)
	return subscriptionsapi.NewDeleteSubscriptionOK().WithPayload(s.ToModel())
}
