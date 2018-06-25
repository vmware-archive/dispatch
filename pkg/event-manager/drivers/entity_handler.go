///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package drivers

import (
	"context"
	"reflect"
	"strings"
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
	defer func() { h.store.Update(ctx, driver.GetRevision(), driver) }()

	// deploy the deployment in k8s cluster

	if err := h.backend.Deploy(ctx, driver); err != nil {
		translateErrorToEntityState(driver, err)
		return ewrapper.Wrap(err, "error deploying driver")
	}

	if driver.Expose {
		log.Infof("Exposing %s event driver", driver.GetName())
		if err := h.backend.Expose(ctx, driver); err != nil {
			translateErrorToEntityState(driver, err)
			return ewrapper.Wrap(err, "error exposing driver")
		}
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
	defer func() { h.store.Update(ctx, driver.GetRevision(), driver) }()

	if err := h.backend.Update(ctx, driver); err != nil {
		translateErrorToEntityState(driver, err)
		return ewrapper.Wrap(err, "error updating driver")
	}

	driver.Status = entitystore.StatusREADY
	driver.SetReason([]string{})

	log.Infof("%s-driver %s has been updated", driver.Type, driver.Name)

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
		translateErrorToEntityState(driver, err)
		h.store.Update(ctx, driver.GetRevision(), driver)
		return ewrapper.Wrap(err, "error deleting driver")
	}

	if err := h.store.Delete(ctx, driver.OrganizationID, driver.Name, driver); err != nil {
		return ewrapper.Wrap(err, "store error when deleting driver")
	}
	log.Infof("driver %s deleted from k8s and the entity store", driver.Name)
	return nil
}

// Sync Executes sync loop
func (h *EntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	// list entity filter
	filter := entitystore.FilterEverything().Add(
		entitystore.FilterStat{
			Scope:   entitystore.FilterScopeField,
			Subject: "ModifiedTime",
			Verb:    entitystore.FilterVerbBefore,
			Object:  time.Now().Add(-resyncPeriod),
		})
	syncingEntities, err := controller.DefaultSync(ctx, h.store, h.Type(), resyncPeriod, filter)
	if err != nil {
		return nil, err
	}

	return syncingEntities, nil
}

// Error handles error state
func (h *EntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	driver := obj.(*entities.Driver)
	var err error
	defer func() { h.store.UpdateWithError(ctx, driver, err) }()

	if len(driver.GetReason()) == 0 {
		return ewrapper.Errorf("%s without error reason", driver.GetName())
	}
	log.Debugf("%s error: reasons are: %s", driver.GetName(), driver.GetReason())

	// TODO: recover/handle error state
	recover := false
	switch driver.GetReason()[0] {
	case errReasonDeploymentNotFound:
		if driver.GetDelete() {
			// in DELETE status, delete driver entity
			log.Debugf("%s in delete state, deployment not found, delete entity")
			h.store.Delete(ctx, driver.OrganizationID, driver.Name, driver)
			recover = true
		}
	case errReasonDeploymentAlreadyExists, errReasonDeploymentNotAvaialble:
		// do update
		h.Update(ctx, driver)
	default:
		log.Debug("other error")
	}

	if recover {
		log.Debugf("%s recovered", driver.Name)
	} else {
		log.Debugf("%s failed to recover: %s", driver.Name, strings.Join(driver.GetReason(), ", "))
	}

	return err
}

func translateErrorToEntityState(driver *entities.Driver, e error) {
	if e == nil {
		return
	}
	log.Debugf("put driver to error state %s: %s", driver.GetName(), e)

	reason := []string{
		e.Error(),
	}
	if c, ok := e.(Causer); ok {
		log.Debugf("%s -- underlying error reason: %s", driver.GetName(), c.Cause().Error())
		reason = append(reason, c.Cause().Error())
	}
	driver.SetReason(reason)
	driver.SetStatus(entitystore.StatusERROR)
}
