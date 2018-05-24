///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// ControllerConfig defines the image manager controller configuration
type ControllerConfig struct {
	ResyncPeriod   time.Duration
	Timeout        time.Duration
	OrganizationID string
}

type baseImageEntityHandler struct {
	Store   entitystore.EntityStore
	Builder *BaseImageBuilder
}

func (h *baseImageEntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&BaseImage{})
}

func (h *baseImageEntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	bi := obj.(*BaseImage)

	defer func() { h.Store.UpdateWithError(ctx, bi, err) }()

	bi.Status = entitystore.StatusREADY
	if err = h.Builder.baseImagePull(ctx, bi); err != nil {
		span.LogKV("error", err)
	}

	return
}

func (h *baseImageEntityHandler) Update(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return h.Add(ctx, obj)
}

func (h *baseImageEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	bi := obj.(*BaseImage)

	err := h.Builder.baseImageDelete(ctx, bi)
	if err != nil {
		span.LogKV("error", err)
		log.Error(err)
	}

	var deleted BaseImage
	err = h.Store.Delete(ctx, bi.GetOrganizationID(), bi.GetName(), &deleted)
	if err != nil {
		return errors.Wrapf(err, "Error deleting base image entity %s/%s", bi.GetOrganizationID(), bi.GetName())
	}
	return nil
}

func (h *baseImageEntityHandler) Sync(ctx context.Context, organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return h.Builder.baseImageStatus(ctx)
}

func (h *baseImageEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	_, err := h.Store.Update(ctx, obj.GetRevision(), obj)
	return err
}

type imageEntityHandler struct {
	Store   entitystore.EntityStore
	Builder *ImageBuilder
	// TODO (bjung): This is a temporary workaround for concurrency issues
	// while building images.  A holistic solution for all controllers is
	// needed.
	lock    sync.Map
	timeout time.Duration
}

// CheckAndLock checks to see if the entity is currently being worked on if not
// stores the entity
func (h *imageEntityHandler) CheckAndLock(obj entitystore.Entity) bool {
	// Only do a single image create at a time (for a given image)
	now := time.Now()
	v, loaded := h.lock.LoadOrStore(obj.GetID(), now)
	if loaded {
		lockTime := v.(time.Time)
		duration := now.Sub(lockTime)
		if duration > h.timeout {
			log.Errorf("Timeout waiting for lock on %s/%s", obj.GetOrganizationID(), obj.GetName())
			// Reset the clock and continue
			h.lock.Store(obj.GetID(), now)
		} else {
			log.Infof("Operation in progress on %s/%s", obj.GetOrganizationID(), obj.GetName())
			return true
		}
	}
	return false
}

// ForceLock grabs the lock regardless of status
func (h *imageEntityHandler) ForceLock(obj entitystore.Entity) {
	h.lock.Store(obj.GetID(), time.Now())
}

// Unlock releases the lock
func (h *imageEntityHandler) Unlock(obj entitystore.Entity) {
	h.lock.Delete(obj.GetID())
}

func (h *imageEntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&Image{})
}

func (h *imageEntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	i := obj.(*Image)

	var bi BaseImage
	err = h.Store.Get(ctx, i.OrganizationID, i.BaseImageName, entitystore.Options{}, &bi)
	if err != nil {
		span.LogKV("error", err)
		log.Error(err)
	}

	if h.CheckAndLock(i) {
		return nil
	}

	defer h.Unlock(i)
	defer func() { h.Store.UpdateWithError(ctx, i, err) }()

	log.Infof("Creating image %s/%s", i.OrganizationID, i.Name)
	if err = h.Builder.imageCreate(ctx, i, &bi); err != nil {
		i.Status = entitystore.StatusERROR
		i.Reason = []string{err.Error()}
		span.LogKV("error", err)
		log.Errorf("Failed to create image %s/%s: %s", i.OrganizationID, i.Name, err)
		return
	}
	log.Infof("Successfully created image %s/%s", i.OrganizationID, i.Name)
	return
}

func (h *imageEntityHandler) Update(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return h.Add(ctx, obj)
}

func (h *imageEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	i := obj.(*Image)

	h.ForceLock(i)
	defer h.Unlock(i)

	err := h.Builder.imageDelete(ctx, i)
	if err != nil {
		span.LogKV("error", err)
		log.Error(err)
	}

	var deleted BaseImage
	err = h.Store.Delete(ctx, i.GetOrganizationID(), i.GetName(), &deleted)
	if err != nil {
		return errors.Wrapf(err, "error deleting image entity %s/%s", i.GetOrganizationID(), i.GetName())
	}
	return nil
}

func (h *imageEntityHandler) Sync(ctx context.Context, organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return h.Builder.imageStatus(ctx)
}

func (h *imageEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	_, err := h.Store.Update(ctx, obj.GetRevision(), obj)
	return err
}

// NewController creates a new image manager controller
func NewController(config *ControllerConfig, store entitystore.EntityStore, baseImageBuilder *BaseImageBuilder, imageBuilder *ImageBuilder) controller.Controller {
	c := controller.NewController(controller.Options{
		OrganizationID: ImageManagerFlags.OrgID,
		ResyncPeriod:   config.ResyncPeriod,
		Workers:        10, // want more functions concurrently? add more workers // TODO configure workers
	})

	c.AddEntityHandler(&baseImageEntityHandler{Store: store, Builder: baseImageBuilder})
	c.AddEntityHandler(&imageEntityHandler{Store: store, Builder: imageBuilder, timeout: config.Timeout})

	return c
}
