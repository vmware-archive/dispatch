///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"time"

	"github.com/casbin/casbin"

	"github.com/vmware/dispatch/pkg/controller"
	"github.com/vmware/dispatch/pkg/entity-store"
)

// NewIdentityController creates a new controller to manage the reconciliation of policy entities
func NewIdentityController(store entitystore.EntityStore, enforcer *casbin.SyncedEnforcer, resync time.Duration) controller.Controller {
	c := controller.NewController(controller.Options{
		ResyncPeriod: resync,
		Workers:      5, // TODO: make this configurable
	})

	c.AddEntityHandler(&policyEntityHandler{store: store, enforcer: enforcer})

	return c
}
