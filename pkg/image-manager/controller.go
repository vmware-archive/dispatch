///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"context"
	"reflect"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// ControllerConfig defines the image manager controller configuration
type ControllerConfig struct {
	ResyncPeriod time.Duration
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
	log.Debugf("Updating last pull time for %s/%s", bi.OrganizationID, bi.Name)
	bi.LastPullTime = time.Now()

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

func (h *baseImageEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
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

	defer func() { h.Store.UpdateWithError(ctx, i, err) }()

	if i.Status == entitystore.StatusMISSING {
		err = h.Builder.imagePull(ctx, i)
	} else {
		err = h.Builder.imageCreate(ctx, i, &bi)
	}

	if err != nil {
		span.LogKV("error", err)
		log.Error(err)
	}
	return
}

func (h *imageEntityHandler) Update(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	i := obj.(*Image)

	if err := h.Builder.imageDelete(ctx, i); err != nil {
		log.Errorf("error deleting docker image %s: %+v", i.DockerURL, err)
	}

	return h.Add(ctx, obj)
}

func (h *imageEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	i := obj.(*Image)

	err := h.Builder.imageDelete(ctx, i)
	if err != nil {
		span.LogKV("error", err)
		log.Error(err)
	}

	var deleted Image
	err = h.Store.Delete(ctx, i.GetOrganizationID(), i.GetName(), &deleted)
	if err != nil {
		return errors.Wrapf(err, "error deleting image entity %s/%s", i.GetOrganizationID(), i.GetName())
	}
	return nil
}

func (h *imageEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
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
		ResyncPeriod:      config.ResyncPeriod,
		Workers:           10, // want more functions concurrently? add more workers // TODO configure workers
		ServiceName:       "images",
		ZookeeperLocation: "transport-zookeeper",
	})

	c.AddEntityHandler(&baseImageEntityHandler{Store: store, Builder: baseImageBuilder})
	c.AddEntityHandler(&imageEntityHandler{Store: store, Builder: imageBuilder})

	return c
}
