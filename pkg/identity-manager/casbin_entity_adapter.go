///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package identitymanager

import (
	"context"
	"fmt"

	casbinModel "github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/vmware/dispatch/pkg/entity-store"
)

// CasbinEntityAdapter type
type CasbinEntityAdapter struct {
	store entitystore.EntityStore
}

// NewCasbinEntityAdapter creates a new casbin entity adapter
func NewCasbinEntityAdapter(store entitystore.EntityStore) *CasbinEntityAdapter {
	return &CasbinEntityAdapter{store: store}
}

// LoadPolicy loads a policy into the casbin entity adapter
func (a *CasbinEntityAdapter) LoadPolicy(model casbinModel.Model) error {
	var policies []*Policy

	// We ignore policy status since the status is only meaningful to the controllers.
	opts := entitystore.Options{
		Filter: entitystore.FilterExists(),
	}

	log.Debug("Reloading policies")
	// The entity adapter loads policies across all orgs into the casbin enforcer. During policy check, the user-specified org-id in header along with other request attributes are validated with the enforcer.
	err := a.store.ListGlobal(context.TODO(), opts, &policies)
	for _, policy := range policies {
		// Casbin authorization rules are of the form (org, subject, resource, action) and hence the need to iterate over all rule fields.
		log.Debugf("Loading policy %s", policy.Name)
		var global string
		if policy.Global {
			global = "y"
		} else {
			global = "n"
		}
		for _, rule := range policy.Rules {
			for _, subject := range rule.Subjects {
				for _, resource := range rule.Resources {
					for _, action := range rule.Actions {
						lineText := fmt.Sprintf("p, %s, %s, %s, %s, %s", global, policy.OrganizationID, subject, resource, action)
						persist.LoadPolicyLine(lineText, model)
					}
				}
			}
		}
	}
	return err
}

// SavePolicy saves all policy rules to the storage.
func (a *CasbinEntityAdapter) SavePolicy(model casbinModel.Model) error {
	return errors.New("not implemented")
}

// AddPolicy adds a policy rule to the storage.
func (a *CasbinEntityAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemovePolicy removes a policy rule from the storage.
func (a *CasbinEntityAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *CasbinEntityAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("not implemented")
}
