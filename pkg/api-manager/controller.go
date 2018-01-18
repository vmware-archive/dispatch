///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package apimanager

import (
	"reflect"
	"time"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/errors"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// ControllerConfig defines configuration for controller
type ControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
}

type apiEntityHandler struct {
	store entitystore.EntityStore
	gw    gateway.Gateway
}

func (h *apiEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&API{})
}

func (h *apiEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Tracef("name %s", obj.GetName())()

	api := obj.(*API)

	defer func() { h.store.UpdateWithError(api, err) }()

	gwAPI, err := h.gw.UpdateAPI(api.Name, &api.API)
	if err != nil {
		return ewrapper.Wrap(err, "gateway error when adding api")
	}
	log.Infof("api %s added by gateway", api.Name)
	api.Status = entitystore.StatusREADY
	api.API.ID = gwAPI.ID
	api.API.CreatedAt = gwAPI.CreatedAt

	return nil
}

func (h *apiEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()

	return h.Add(obj)
}

func (h *apiEntityHandler) Delete(obj entitystore.Entity) error {
	defer trace.Tracef("name '%s'", obj.GetName())()

	api, ok := obj.(*API)
	if !ok {
		return ewrapper.New("type assertion error")
	}
	if err := h.gw.DeleteAPI(&api.API); err != nil {
		if _, ok := err.(*errors.ObjectNotFoundError); !ok {
			return ewrapper.Wrap(err, "gateway error when deleting api")
		}
		// object not found, continue to delete from entity store
	}

	// TODO: consider to keep one of the below
	// soft deletion, when UUID is used
	// api.Status = entitystore.StatusDELETED
	// if _, err := store.Update(api.Revision, api); err != nil {
	// 	log.Errorf("store error when updating api: %+v", err)
	// }

	// hard deletion
	if err := h.store.Delete(api.OrganizationID, api.Name, api); err != nil {
		return ewrapper.Wrap(err, "store error when deleting api")
	}
	log.Infof("api %s deleted by gateway and store", api.Name)
	return nil
}

func (h *apiEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	return controller.DefaultSync(h.store, h.Type(), organizationID, resyncPeriod)
}

func (h *apiEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Tracef("")()

	api, ok := obj.(*API)
	if !ok {
		return ewrapper.New("type assertion error")
	}

	// delete the underlying api
	h.gw.DeleteAPI(&api.API)

	// try to update api again
	return h.Update(api)
}

// NewController creates a new controller
func NewController(config *ControllerConfig, store entitystore.EntityStore, gw gateway.Gateway) controller.Controller {
	defer trace.Trace("")()

	c := controller.NewController(controller.Options{
		OrganizationID: config.OrganizationID,
		ResyncPeriod:   config.ResyncPeriod,
	})

	c.AddEntityHandler(&apiEntityHandler{store: store, gw: gw})
	return c
}
