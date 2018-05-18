///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import (
	"context"
	"reflect"
	"time"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/event-manager/drivers/entities"
	"github.com/vmware/dispatch/pkg/trace"
)

// EntityHandler handles driver entity operations
type EntityHandler struct {
	store   entitystore.EntityStore
	backend Backend
}

// NewEntityHandler creates new instance of EntityHandler
func NewEntityHandler(store entitystore.EntityStore, backend Backend) *EntityHandler {
	return &EntityHandler{
		store:   store,
		backend: backend,
	}
}

// Type returns Entity Handler type
func (h *EntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&entities.Driver{})
}

// Add adds new driver to the store, and executes its deployment.
func (h *EntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	driver := obj.(*entities.Driver)
	defer func() { h.store.UpdateWithError(ctx, driver, err) }()

	// deploy the deployment in k8s cluster

	if err := h.backend.Deploy(ctx, driver); err != nil {
		return ewrapper.Wrap(err, "error deploying driver")
	}

	driver.Status = entitystore.StatusREADY

	log.Infof("%s-driver %s has been deployed on k8s", driver.Type, driver.Name)

	return nil
}

// Update updates the driver by updating the deployment
func (h *EntityHandler) Update(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	driver := obj.(*entities.Driver)
	defer func() { h.store.UpdateWithError(ctx, driver, err) }()

	if err := h.backend.Update(ctx, driver); err != nil {
		return ewrapper.Wrap(err, "error updating driver")
	}

	driver.Status = entitystore.StatusREADY

	log.Info("%s-driver %s has been updated", driver.Type, driver.Name)

	return nil
}

// Delete deletes the driver from the backend
func (h *EntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	driver := obj.(*entities.Driver)

	// delete the deployment from k8s cluster
	err := h.backend.Delete(ctx, driver)
	if err != nil {
		return ewrapper.Wrap(err, "error deleting driver")
	}

	if err := h.store.Delete(ctx, driver.OrganizationID, driver.Name, driver); err != nil {
		return ewrapper.Wrap(err, "store error when deleting driver")
	}
	log.Infof("driver %s deleted from k8s and the entity store", driver.Name)
	return nil
}

// Sync Executes sync loop
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
