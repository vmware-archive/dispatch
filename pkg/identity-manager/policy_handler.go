///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"reflect"
	"time"

	"github.com/casbin/casbin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
	"github.com/vmware/dispatch/pkg/trace"
)

type policyEntityHandler struct {
	store    entitystore.EntityStore
	enforcer *casbin.SyncedEnforcer
}

func (h *policyEntityHandler) Type() reflect.Type {
	defer trace.Trace("")()

	return reflect.TypeOf(&Policy{})
}

func (h *policyEntityHandler) Add(obj entitystore.Entity) (err error) {
	defer trace.Tracef("name %s", obj.GetName())()

	policy := obj.(*Policy)
	defer func() { h.store.UpdateWithError(policy, err) }()

	policy.Status = entitystore.StatusREADY

	if err := h.enforcer.LoadPolicy(); err != nil {
		return errors.Wrap(err, "error when re-loading policies")
	}

	log.Infof("policy %s has been created", policy.Name)

	return nil
}

func (h *policyEntityHandler) Update(obj entitystore.Entity) error {
	defer trace.Trace("")()
	return h.Add(obj)
}

func (h *policyEntityHandler) Delete(obj entitystore.Entity) (err error) {
	defer trace.Tracef("name '%s'", obj.GetName())()

	policy := obj.(*Policy)

	// hard deletion
	if err := h.store.Delete(policy.OrganizationID, policy.Name, policy); err != nil {
		return errors.Wrap(err, "store error when deleting policy")
	}

	if err := h.enforcer.LoadPolicy(); err != nil {
		return errors.Wrap(err, "error when re-loading policies")
	}

	log.Infof("policy %s deleted from the entity store", policy.Name)
	return nil
}

func (h *policyEntityHandler) Sync(organizationID string, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	defer trace.Trace("")()
	// TODO: Move this out of entity handler sync to controller's
	// Reload policies
	if err := h.enforcer.LoadPolicy(); err != nil {
		return nil, errors.Wrap(err, "error when re-loading policies")
	}
	return controller.DefaultSync(h.store, h.Type(), organizationID, resyncPeriod)
}

func (h *policyEntityHandler) Error(obj entitystore.Entity) error {
	defer trace.Tracef("")()

	log.Errorf("handleError func not implemented yet")
	return nil
}
