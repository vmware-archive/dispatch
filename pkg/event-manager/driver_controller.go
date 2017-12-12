///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

// NO TEST

import (
	"time"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"

	entitystore "github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// EventDriverControllerConfig defines configuration for controller
type EventDriverControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
}

// NewEventDriverController creates a new controller
func NewEventDriverController(config *EventDriverControllerConfig, lw controller.ListerWatcher, store entitystore.EntityStore) (controller.Controller, error) {
	defer trace.Trace("")()

	k8sHelper, err := NewK8sHelper()
	if err != nil {
		log.Errorf("error creating k8s helper: %+v", err)
	}

	addFunc := func(obj entitystore.Entity) error {
		defer trace.Tracef("name %s", obj.GetName())()

		driver, ok := obj.(*Driver)
		if !ok {
			log.Errorf("type assertion error")
		}

		// deploy the deployment in k8s cluster
		err := k8sHelper.Deploy(driver)
		if err != nil {
			err = ewrapper.Wrap(err, "error deploying driver")
			log.Errorln(err)
			return err
		}

		driver.Status = entitystore.StatusREADY
		if _, err := store.Update(driver.Revision, driver); err != nil {
			log.Errorf("store error when updating event driver: %+v", err)
		}
		log.Infof("%s-driver %s has been deployed on k8s", driver.Type, driver.Name)
		return nil
	}

	deleteFunc := func(obj entitystore.Entity) error {
		defer trace.Tracef("name '%s'", obj.GetName())()
		driver, ok := obj.(*Driver)
		if !ok {
			log.Errorf("type assertion error")
		}

		// delete the deployment from k8s cluster
		err := k8sHelper.Delete(driver)
		if err != nil {
			err = ewrapper.Wrap(err, "error deleting driver")
			log.Errorln(err)
			return err
		}

		// TODO: consider to keep one of the below
		// soft deletion, when UUID is used
		// driver.Status = entitystore.StatusDELETED
		// if _, err := store.Update(driver.Revision, driver); err != nil {
		// 	log.Errorf("store error when updating driver: %+v", err)
		// }

		// hard deletion
		if err := store.Delete(driver.OrganizationID, driver.Name, driver); err != nil {
			err = ewrapper.Wrap(err, "store error when deleting api")
			log.Error(err)
			return err
		}
		log.Infof("driver %s deleted from k8s and the entity store", driver.Name)
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
