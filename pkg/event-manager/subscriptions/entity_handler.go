///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package subscriptions

import (
	"context"
	"reflect"
	"time"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/subscriptions/entities"
	"github.com/vmware/dispatch/pkg/trace"
)

// EntityHandler handles Subscription entity operations
type EntityHandler struct {
	store   entitystore.EntityStore
	manager Manager
}

// NewEntityHandler returns new instance of EntityHandler
func NewEntityHandler(store entitystore.EntityStore, manager Manager) *EntityHandler {
	return &EntityHandler{
		store:   store,
		manager: manager,
	}
}

// Type returns entity handler type
func (h *EntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&entities.Subscription{})
}

// Add handles adding new subscription entity
func (h *EntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	sub := obj.(*entities.Subscription)
	defer func() { h.store.UpdateWithError(ctx, sub, err) }()

	if err := h.manager.Create(context.Background(), sub); err != nil {
		return ewrapper.Wrap(err, "error activating subscription")
	}

	sub.Status = entitystore.StatusREADY

	log.Infof("subscription %s for event type %s has been activated", sub.Name, sub.EventType)

	return nil
}

// Update handles subscription entity update
func (h *EntityHandler) Update(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	sub := obj.(*entities.Subscription)
	defer func() { h.store.UpdateWithError(ctx, sub, err) }()

	if err := h.manager.Update(context.Background(), sub); err != nil {
		return ewrapper.Wrap(err, "error activating subscription")
	}

	sub.Status = entitystore.StatusREADY

	log.Infof("subscription %s for event type %s has been updated", sub.Name, sub.EventType)
	return h.Add(ctx, obj)
}

// Delete handles subscription entity deletion
func (h *EntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	sub := obj.(*entities.Subscription)

	// unsubscribe from queue
	err := h.manager.Delete(context.Background(), sub)
	if err != nil {
		return ewrapper.Wrap(err, "error deactivating subscription")
	}

	// hard deletion
	if err := h.store.Delete(ctx, sub.OrganizationID, sub.Name, sub); err != nil {
		return ewrapper.Wrap(err, "store error when deleting subscription")
	}
	log.Infof("subscription %s deactivated and deleted from the entity store", sub.Name)
	return nil
}

// Sync is responsible for syncing the state of active subscriptions and their entities
func (h *EntityHandler) Sync(ctx context.Context, organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return controller.DefaultSync(ctx, h.store, h.Type(), organizationID, resyncPeriod, nil)
}

// Error handles error state
func (h *EntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	log.Errorf("handleError func not implemented yet")
	return nil
}
