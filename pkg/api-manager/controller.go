///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package apimanager

import (
	"time"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/api-manager/gateway"
	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/errors"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// ControllerConfig defines configuration for controller
type ControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
}

// NewAPIController creates a new controller
func NewAPIController(config *ControllerConfig, lw controller.ListerWatcher, store entitystore.EntityStore, gw gateway.Gateway) (controller.Controller, error) {
	defer trace.Trace("")()

	addFunc := func(obj entitystore.Entity) error {
		defer trace.Tracef("name %s", obj.GetName())()

		api, ok := obj.(*API)
		if !ok {
			log.Errorf("type assertion error")
		}
		gwAPI, err := gw.UpdateAPI(api.Name, &api.API)
		if err != nil {
			err = ewrapper.Wrap(err, "gateway error when adding api")
			log.Error(err)
			return err
		}
		api.Status = entitystore.StatusREADY
		api.API.ID = gwAPI.ID
		api.API.CreatedAt = gwAPI.CreatedAt
		if _, err := store.Update(api.Revision, api); err != nil {
			log.Errorf("store error when updating api: %+v", err)
		}
		log.Infof("api %s added by gateway", api.Name)
		return nil
	}

	deleteFunc := func(obj entitystore.Entity) error {
		defer trace.Tracef("name '%s'", obj.GetName())()
		api, ok := obj.(*API)
		if !ok {
			log.Errorf("type assertion error")
		}
		if err := gw.DeleteAPI(&api.API); err != nil {
			if _, ok := err.(*errors.ObjectNotFoundError); !ok {
				err = ewrapper.Wrap(err, "gateway error when deleting api")
				log.Error(err)
				return err
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
		if err := store.Delete(api.OrganizationID, api.Name, api); err != nil {
			err = ewrapper.Wrap(err, "store error when deleting api")
			log.Error(err)
			return err
		}
		log.Infof("api %s deleted by gateway and store", api.Name)
		return nil
	}

	handlers := controller.EventHandlers{
		AddFunc:    addFunc,
		UpdateFunc: func(_, obj entitystore.Entity) error { return addFunc(obj) },
		DeleteFunc: deleteFunc,
		ErrorFunc: func(obj entitystore.Entity) error {
			defer trace.Tracef("")()
			log.Errorf("handleError func not implemented yet")
			return nil
		},
	}
	informer := controller.NewDefaultInformer(lw, config.ResyncPeriod)
	informer.AddEventHandlers(handlers)
	return controller.NewDefaultController(informer)
}
