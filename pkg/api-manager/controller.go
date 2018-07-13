///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package apimanager

import (
	"context"
	"reflect"
	"time"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/errors"
	"github.com/vmware/dispatch/pkg/trace"
)

// ControllerConfig defines configuration for controller
type ControllerConfig struct {
	ResyncPeriod time.Duration
}

type apiEntityHandler struct {
	store entitystore.EntityStore
	gw    gateway.Gateway
}

func (h *apiEntityHandler) Type() reflect.Type {
	return reflect.TypeOf(&API{})
}

// Add is the handler for creating API endpoints
func (h *apiEntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	api := obj.(*API)

	defer func() { h.store.UpdateWithError(ctx, api, err) }()

	gwAPI, err := h.gw.AddAPI(ctx, &api.API)
	if err != nil {
		return ewrapper.Wrap(err, "gateway error when adding api")
	}
	log.Infof("api %s added in gateway", api.API.Name)
	api.Status = entitystore.StatusREADY
	api.API.ID = gwAPI.ID
	api.API.CreatedAt = gwAPI.CreatedAt

	return nil
}

// Update is the handler for updating API endpoints
func (h *apiEntityHandler) Update(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	api := obj.(*API)

	defer func() { h.store.UpdateWithError(ctx, api, err) }()

	gwAPI, err := h.gw.UpdateAPI(ctx, api.API.Name, &api.API)
	if err != nil {
		return ewrapper.Wrap(err, "gateway error when updating api")
	}
	log.Infof("api %s updated in gateway", api.API.Name)
	api.Status = entitystore.StatusREADY
	api.API.ID = gwAPI.ID

	return nil
}

// Delete is the handler for deleting API endpoints
func (h *apiEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	api, ok := obj.(*API)
	if !ok {
		return ewrapper.New("type assertion error")
	}
	if err := h.gw.DeleteAPI(ctx, &api.API); err != nil {
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
	if err := h.store.Delete(ctx, api.OrganizationID, api.Name, api); err != nil {
		return ewrapper.Wrap(err, "store error when deleting api")
	}
	log.Infof("api %s deleted by gateway and store", api.Name)
	return nil
}

// Sync polls the actual state and returns the list of entites which need to be resolved
func (h *apiEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return controller.DefaultSync(ctx, h.store, h.Type(), resyncPeriod, nil)
}

// Error handles errors while modifying API endpoints
func (h *apiEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	api, ok := obj.(*API)
	if !ok {
		return ewrapper.New("type assertion error")
	}

	// delete the underlying api
	err := h.gw.DeleteAPI(ctx, &api.API)
	if err != nil {
		return ewrapper.Wrap(err, "kong error deleting API")
	}

	// try to update api again
	return h.Update(ctx, api)
}

// NewController creates a new controller
func NewController(config *ControllerConfig, store entitystore.EntityStore, gw gateway.Gateway) controller.Controller {
	c := controller.NewController(controller.Options{
		ServiceName:       "APIs",
		ResyncPeriod:      config.ResyncPeriod,
		ZookeeperLocation: "transport-zookeeper",
	})

	c.AddEntityHandler(&apiEntityHandler{store: store, gw: gw})
	return c
}
