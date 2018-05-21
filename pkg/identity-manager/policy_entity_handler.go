///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"context"
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
	return reflect.TypeOf(&Policy{})
}

func (h *policyEntityHandler) Add(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	policy := obj.(*Policy)
	defer func() { h.store.UpdateWithError(ctx, policy, err) }()

	policy.Status = entitystore.StatusREADY

	if err := h.enforcer.LoadPolicy(); err != nil {
		return errors.Wrap(err, "error when re-loading policies")
	}

	log.Infof("policy %s has been created", policy.Name)

	return nil
}

func (h *policyEntityHandler) Update(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	return h.Add(ctx, obj)
}

func (h *policyEntityHandler) Delete(ctx context.Context, obj entitystore.Entity) (err error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	policy := obj.(*Policy)

	// hard deletion
	if err := h.store.Delete(ctx, policy.OrganizationID, policy.Name, policy); err != nil {
		return errors.Wrap(err, "store error when deleting policy")
	}

	if err := h.enforcer.LoadPolicy(); err != nil {
		return errors.Wrap(err, "error when re-loading policies")
	}

	log.Infof("policy %s deleted from the entity store", policy.Name)
	return nil
}

func (h *policyEntityHandler) Sync(ctx context.Context, resyncPeriod time.Duration) ([]entitystore.Entity, error) {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	// TODO: Move this out of entity handler sync to controller's
	// Reload policies
	if err := h.enforcer.LoadPolicy(); err != nil {
		return nil, errors.Wrap(err, "error when re-loading policies")
	}
	return controller.DefaultSync(ctx, h.store, h.Type(), resyncPeriod, nil)
}

func (h *policyEntityHandler) Error(ctx context.Context, obj entitystore.Entity) error {
	span, ctx := trace.Trace(ctx, "")
	defer span.Finish()

	log.Errorf("handleError func not implemented yet")
	return nil
}
