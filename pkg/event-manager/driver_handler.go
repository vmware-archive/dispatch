///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package eventmanager

import (
	"reflect"
	"time"

	ewrapper "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

type driverEntityHandler struct {
	store         entitystore.EntityStore
	driverBackend DriverBackend
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

	if err := h.driverBackend.Deploy(driver); err != nil {
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
	err := h.driverBackend.Delete(driver)
	if err != nil {
		return ewrapper.Wrap(err, "error deleting driver")
	}

	if err := h.store.Delete(driver.OrganizationID, driver.Name, driver); err != nil {
		return ewrapper.Wrap(err, "store error when deleting driver")
	}
	log.Infof("driver %s deleted from k8s and the entity store", driver.Name)
	return nil
}

func (h *driverEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()

	return controller.DefaultSync(h.store, h.Type(), organizationID, resyncPeriod)
}

func (h *driverEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Tracef("")()

	log.Errorf("handleError func not implemented yet")
	return nil
}
