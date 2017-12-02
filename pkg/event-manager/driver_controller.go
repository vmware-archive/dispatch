///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

// NO TEST

import (
	"reflect"
	"time"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"

	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

// EventDriverControllerConfig defines configuration for controller
type EventDriverControllerConfig struct {
	ResyncPeriod   time.Duration
	OrganizationID string
}

type driverEntityHandler struct {
	store     entitystore.EntityStore
	k8sHelper *K8sHelper
}

func (h *driverEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&Driver{})
}

func (h *driverEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Tracef("name %s", obj.GetName())()

	driver := obj.(*Driver)
	defer func() { h.store.UpdateWithError(driver, err) }()

	// deploy the deployment in k8s cluster

	if err := h.k8sHelper.Deploy(driver); err != nil {
		return ewrapper.Wrap(err, "error deploying driver")
	}

	driver.Status = entitystore.StatusREADY

	log.Infof("%s-driver %s has been deployed on k8s", driver.Type, driver.Name)

	return nil
}

func (h *driverEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()

	return h.Add(obj)
}

func (h *driverEntityHandler) Delete(obj entitystore.Entity) error {
	defer trace.Tracef("name '%s'", obj.GetName())()

	driver := obj.(*Driver)

	// delete the deployment from k8s cluster
	err := h.k8sHelper.Delete(driver)
	if err != nil {
		return ewrapper.Wrap(err, "error deleting driver")
	}

	// TODO: consider to keep one of the below
	// soft deletion, when UUID is used
	// driver.Status = entitystore.StatusDELETED
	// if _, err := store.Update(driver.Revision, driver); err != nil {
	// 	log.Errorf("store error when updating driver: %+v", err)
	// }

	// hard deletion
	if err := h.store.Delete(driver.OrganizationID, driver.Name, driver); err != nil {
		return ewrapper.Wrap(err, "store error when deleting api")
	}
	log.Infof("driver %s deleted from k8s and the entity store", driver.Name)
	return nil
}

func (h *driverEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Tracef("")()

	log.Errorf("handleError func not implemented yet")
	return nil
}

// NewEventDriverController creates a new controller
func NewEventDriverController(config *EventDriverControllerConfig, store entitystore.EntityStore) controller.Controller {
	defer trace.Trace("")()

	k8sHelper, err := NewK8sHelper()
	if err != nil {
		log.Errorf("error creating k8s helper: %+v", err)
	}

	c := controller.NewController(store, controller.Options{
		OrganizationID: config.OrganizationID,
		ResyncPeriod:   config.ResyncPeriod,
		Workers:        1000, // just add more if you need more
	})

	c.AddEntityHandler(&driverEntityHandler{store: store, k8sHelper: k8sHelper})

	return c
}
