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
	"github.com/vmware/dispatch/pkg/trace"
)

// NewIdentityController creates a new controller to manage the reconciliation of policy entities
func NewIdentityController(store entitystore.EntityStore, enforcer *casbin.SyncedEnforcer) controller.Controller {
	defer trace.Trace("")()

	c := controller.NewController(controller.Options{
		OrganizationID: IdentityManagerFlags.OrgID,
		ResyncPeriod:   time.Duration(IdentityManagerFlags.ResyncPeriod) * time.Second,
		Workers:        5, // TODO: make this configurable
	})

	c.AddEntityHandler(&policyEntityHandler{store: store, enforcer: enforcer})

	return c
}
