///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package imagemanager

import (
	"reflect"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

type ControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
}

type baseImageEntityHandler struct {
	Store   entitystore.EntityStore
	Builder *BaseImageBuilder
}

func (h *baseImageEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&BaseImage{})
}

func (h *baseImageEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	bi := obj.(*BaseImage)

	defer func() { h.Store.UpdateWithError(bi, err) }()

	bi.Status = entitystore.StatusREADY
	if err := h.Builder.baseImagePull(bi); err != nil {
		bi.Status = entitystore.StatusERROR
		bi.Reason = []string{err.Error()}
	}

	return
}

func (h *baseImageEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()

	return h.Add(obj)
}

func (h *baseImageEntityHandler) Delete(obj entitystore.Entity) error {
	defer trace.Trace("")()

	bi := obj.(*BaseImage)

	err := h.Builder.baseImageDelete(bi)
	if err != nil {
		log.Error(err)
	}

	var deleted BaseImage
	err = h.Store.Delete(bi.GetOrganizationID(), bi.GetName(), &deleted)
	if err != nil {
		return errors.Wrapf(err, "Error deleting base image entity %s/%s", bi.GetOrganizationID(), bi.GetName())
	}
	return nil
}

func (h *baseImageEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	return h.Builder.baseImageStatus()
}

func (h *baseImageEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Trace("")()

	_, err := h.Store.Update(obj.GetRevision(), obj)
	return err
}

type imageEntityHandler struct {
	Store   entitystore.EntityStore
	Builder *ImageBuilder
}

func (h *imageEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&Image{})
}

func (h *imageEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Trace("")()

	i := obj.(*Image)

	var bi BaseImage
	err = h.Store.Get(i.OrganizationID, i.BaseImageName, entitystore.Options{}, &bi)
	if err != nil {
		i.Status = entitystore.StatusERROR
		i.Reason = []string{err.Error()}
	}

	defer func() { h.Store.UpdateWithError(i, err) }()

	if err := h.Builder.imageCreate(i, &bi); err != nil {
		i.Status = entitystore.StatusERROR
		i.Reason = []string{err.Error()}
	}
	return
}

func (h *imageEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()

	return h.Add(obj)
}

func (h *imageEntityHandler) Delete(obj entitystore.Entity) error {
	defer trace.Trace("")()

	i := obj.(*Image)

	err := h.Builder.imageDelete(i)
	if err != nil {
		log.Error(err)
	}

	var deleted BaseImage
	err = h.Store.Delete(i.GetOrganizationID(), i.GetName(), &deleted)
	if err != nil {
		err = errors.Wrapf(err, "error deleting image entity %s/%s", i.GetOrganizationID(), i.GetName())
		log.Error(err)
		return err
	}
	return nil
}

func (h *imageEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	return h.Builder.imageStatus()
}

func (h *imageEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Trace("")()

	_, err := h.Store.Update(obj.GetRevision(), obj)
	return err
}

func NewController(config *ControllerConfig, store entitystore.EntityStore, baseImageBuilder *BaseImageBuilder, imageBuilder *ImageBuilder) controller.Controller {

	defer trace.Trace("")()

	c := controller.NewController(controller.Options{
		OrganizationID: ImageManagerFlags.OrgID,
		ResyncPeriod:   config.ResyncPeriod,
		Workers:        10, // want more functions concurrently? add more workers // TODO configure workers
	})

	c.AddEntityHandler(&baseImageEntityHandler{Store: store, Builder: baseImageBuilder})
	c.AddEntityHandler(&imageEntityHandler{Store: store, Builder: imageBuilder})

	return c
}
